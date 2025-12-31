# AI Chat Bugs Fixes & Quota Refactoring Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix critical bugs in chat message persistence (duplicate history) and refactor quota management for robust daily resets and Redis synchronization.

**Architecture:**
1.  **Message Persistence:** Refactor `saveToSession` to explicitly accept only new messages, decoupling history merging from persistence.
2.  **Quota Management:** Centralize quota logic in `quota.go`. Implement "Self-Healing" `ReserveTPD` that initializes from DB if Redis key is missing, and ensures daily quota reset logic is executed.

**Tech Stack:** Go, GORM, Redis (Go-Redis), Gin

---

## Task 1: Fix Duplicate Message Persistence

**Files:**
- Modify: `internal/apiserver/biz/chat/chat.go`
- Test: `internal/apiserver/biz/chat/chat_test.go` (create if missing)

**Step 1: Create failing test for duplicate messages**

Create `internal/apiserver/biz/chat/chat_test.go` to simulate a chat flow where history is loaded, and verify that `saveToSession` would save duplicate messages with the current implementation (mental check or actual test). Since we are modifying private methods/logic, we'll verify via the `Chat` method behavior or unit test internal methods if exported.

*Note: For this plan, we will focus on the implementation fix as the bug is logic-obvious.*

**Step 2: Refactor `saveToSession` signature**

Update `saveToSession` to accept `newMessages []ai.Message` instead of extracting them from `req.Messages` (which currently contains history).

```go
// internal/apiserver/biz/chat/chat.go

// Old signature:
// func (b *chatBiz) saveToSession(ctx context.Context, uid string, req *ai.ChatRequest, resp *ai.ChatResponse)

// New signature:
func (b *chatBiz) saveToSession(ctx context.Context, uid string, sessionID string, newMessages []ai.Message, resp *ai.ChatResponse) {
    // Save user messages (only the new ones passed in)
    for _, msg := range newMessages {
        if msg.Role == ai.RoleUser {
            if err := b.ds.AiMessage().Create(ctx, &model.AiMessageM{
                SessionID: sessionID,
                Role:      msg.Role,
                Content:   msg.Content,
                Model:     resp.Model, // User message model matches response model
            }); err != nil {
                log.C(ctx).Errorw("Failed to save user message", "session_id", sessionID, "uid", uid, "err", err)
            }
        }
    }
    
    // ... rest of the function (save assistant response)
}
```

**Step 3: Update `Chat` method to capture new messages**

Modify `Chat` method to store original messages before merging history.

```go
// internal/apiserver/biz/chat/chat.go

func (b *chatBiz) Chat(ctx context.Context, uid string, req *ai.ChatRequest) (*ai.ChatResponse, error) {
    // ... validation ...

    // Capture new messages BEFORE loading history
    newMessages := req.Messages

    // Load and merge history messages
    // ... (existing code that modifies req.Messages) ...

    // ... (quota reservation) ...
    // ... (provider call) ...

    // Save to session using captured newMessages
    if req.SessionID != "" {
        go func() {
            ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
            defer cancel()
            // Pass newMessages explicitly
            b.saveToSession(ctx, uid, req.SessionID, newMessages, resp)
        }()
    }

    return resp, nil
}
```

**Step 4: Update `ChatStream` and `saveStreamToSession`**

Similarly update streaming logic.

```go
// internal/apiserver/biz/chat/chat.go

func (b *chatBiz) ChatStream(...) {
    // ...
    newMessages := req.Messages
    // ... load history ...
    
    // Pass newMessages to wrapStreamForSaving
    return b.wrapStreamForSaving(stream, uid, req, newMessages, reservedTokens), nil
}

func (b *chatBiz) wrapStreamForSaving(stream *ai.ChatStream, uid string, req *ai.ChatRequest, newMessages []ai.Message, reservedTokens int) *ai.ChatStream {
    // ...
    // On stream end:
    b.saveStreamToSession(ctx, uid, req.SessionID, newMessages, string(contentBuilder), modelName, totalTokens)
    // ...
}

func (b *chatBiz) saveStreamToSession(ctx context.Context, uid string, sessionID string, newMessages []ai.Message, content string, modelName string, tokens int) {
    // Save user messages (iterate over newMessages)
    for _, msg := range newMessages {
         // ... create db record ...
    }
    // ... save assistant message ...
}
```

**Step 5: Verify build**

Run: `make build`

**Step 6: Commit**

```bash
git add internal/apiserver/biz/chat/chat.go
git commit -m "fix(ai): prevent duplicate history saving in chat session"
```

---

## Task 2: Refactor Quota Management (Self-Healing)

**Files:**
- Modify: `internal/apiserver/biz/chat/quota.go`

**Step 1: Remove unused `CheckTPD` and impl `ensureQuotaState`**

Refactor `quota.go` to include a helper that ensures Redis state is correct and daily reset has happened.

```go
// internal/apiserver/biz/chat/quota.go

// ensureQuotaState ensures the user's daily quota is reset in DB if needed,
// and returns the current used tokens from DB for Redis initialization.
func (q *quotaChecker) ensureQuotaState(ctx context.Context, uid string) (int, int, error) {
    quota, tpd, err := q.getUserQuota(ctx, uid)
    if err != nil {
        return 0, 0, err
    }

    // Check reset
    if q.shouldResetDaily(quota) {
        if err := q.ds.AiUserQuota().ResetDailyTokens(ctx, uid); err != nil {
            return 0, 0, errno.ErrOperationFailed.WithMessage("failed to reset daily quota: %v", err)
        }
        quota.UsedTokensToday = 0
    }
    
    return quota.UsedTokensToday, tpd, nil
}
```

**Step 2: Update `ReserveTPD` to be self-healing**

```go
// internal/apiserver/biz/chat/quota.go

func (q *quotaChecker) ReserveTPD(ctx context.Context, uid string, estimatedTokens int) (int, error) {
    if !facade.Config.AI.Quota.Enabled {
        return 0, nil
    }
    if estimatedTokens <= 0 {
        estimatedTokens = defaultEstimatedTokens
    }

    key := q.buildQuotaKey(uid)

    // ... inside ReserveTPD ...
    
    var tpd int
    var err error

    // 1. Check/Init Redis
    exists, err := facade.Redis.Exists(ctx, key).Result()
    if err != nil {
        return 0, errno.ErrOperationFailed.WithMessage("redis error: %v", err)
    }

    if exists == 0 {
        // Key missing: Initialize from DB
        var usedInDB int
        // Reuse tpd from this call
        usedInDB, tpd, err = q.ensureQuotaState(ctx, uid)
        if err != nil {
            return 0, err
        }
        
        // Atomically set initial value if not exists (NX)
        _, err := facade.Redis.SetNX(ctx, key, usedInDB, quotaKeyTTL).Result()
        if err != nil {
             return 0, errno.ErrOperationFailed.WithMessage("redis setnx error: %v", err)
        }
        
        // Optimistic check: if we just initialized, check limit immediately before incrementing
        if usedInDB+estimatedTokens > tpd {
             return 0, errno.ErrAIQuotaExceeded
        }
    }

    // 2. Increment
    newTotal, err := facade.Redis.IncrBy(ctx, key, int64(estimatedTokens)).Result()
    if err != nil {
        return 0, err
    }
    // Refresh TTL
    facade.Redis.Expire(ctx, key, quotaKeyTTL)

    // 3. Check Limit
    // Only fetch TPD if we haven't already (i.e., Redis key existed)
    if tpd == 0 {
        _, tpd, err = q.getUserQuota(ctx, uid) 
        if err != nil {
             // Rollback
             facade.Redis.DecrBy(ctx, key, int64(estimatedTokens))
             return 0, err
        }
    }

    if int(newTotal) > tpd {
        facade.Redis.DecrBy(ctx, key, int64(estimatedTokens))
        return 0, errno.ErrAIQuotaExceeded.WithMessage("daily token quota exceeded (%d/%d)", int(newTotal)-estimatedTokens, tpd)
    }

    return estimatedTokens, nil
}
```

*Refinement on Step 2:*
Calling `getUserQuota` every time might be acceptable if DB latency is low, but `ensureQuotaState` (checking daily reset) involves DB write on reset day.
To be robust and correct:
1. `getUserQuota` fetches DB record.
2. If `shouldResetDaily`, do reset.
This logic is fine.

**Step 3: Clean up `CheckTPD`**

Delete the deprecated `CheckTPD` and `UpdateTPD` methods if no longer used (checked in Task 1 code, they are not used).

**Step 4: Verify build**

Run: `make build`

**Step 5: Commit**

```bash
git add internal/apiserver/biz/chat/quota.go
git commit -m "fix(ai): implement self-healing quota reservation with db sync"
```

---

## Task 3: Final Verification

**Step 1: Run Lint**

Run: `make lint`

**Step 2: Run Tests**

Run: `go test ./internal/apiserver/biz/chat/...`

**Step 3: Cleanup**

Remove any temporary test files if created.
