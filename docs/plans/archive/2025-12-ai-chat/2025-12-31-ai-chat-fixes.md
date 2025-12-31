# AI Chat Fixes Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix critical data consistency issues, potential quota leaks, and improve code maintainability in the AI Chat module.

**Architecture:** Refactor common provider logic into a shared utility, standardize provider interfaces, and correct the chat business logic for message persistence and quota management.

**Tech Stack:** Go, Gorm (MySQL), Redis (Quota), CloudWeGo/Eino (AI Providers).

---

### Task 1: Extract Common Provider Utilities

**Files:**
- Create: `pkg/ai/common.go`
- Create: `pkg/ai/common_test.go`
- Modify: `pkg/ai/providers/openai/provider.go`
- Modify: `pkg/ai/providers/claude/provider.go`
- Modify: `pkg/ai/providers/gemini/provider.go`
- Modify: `pkg/ai/providers/qwen/provider.go`

**Step 1: Analyze duplicated functions**

After code review, the following functions are **identical** across all 4 providers (OpenAI, Claude, Gemini, Qwen):
- `convertMessages` - Converts `[]ai.Message` to `[]*schema.Message`
- `convertResponse` - Converts `schema.Message` to `ai.ChatResponse`
- `convertStreamChunk` - Converts stream chunk
- `generateID` - Generates unique IDs
- `extractUsage` - Extracts token usage

All providers use the same Eino types (`schema.Message`), so these can be unified.

**Step 2: Create common.go with all utilities**

```go
// ABOUTME: Common utilities for AI providers.
// ABOUTME: Shared conversion and helper functions used by all providers.

package ai

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/cloudwego/eino/schema"
)

// GenerateID generates a unique ID for chat completions.
func GenerateID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "chatcmpl-" + hex.EncodeToString(b)
}

// ConvertMessages converts ai.Message to schema.Message.
func ConvertMessages(msgs []Message) []*schema.Message {
	result := make([]*schema.Message, len(msgs))
	for i, m := range msgs {
		role := schema.User
		switch m.Role {
		case RoleSystem:
			role = schema.System
		case RoleAssistant:
			role = schema.Assistant
		}
		result[i] = &schema.Message{
			Role:    role,
			Content: m.Content,
		}
	}
	return result
}

// ConvertResponse converts Eino response to ai.ChatResponse.
func ConvertResponse(resp *schema.Message, modelName string) *ChatResponse {
	usage := ExtractUsage(resp)

	return &ChatResponse{
		ID:      GenerateID(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   modelName,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    RoleAssistant,
					Content: resp.Content,
				},
				FinishReason: "stop",
			},
		},
		Usage: usage,
	}
}

// ConvertStreamChunk converts Eino stream message to ai.StreamChunk.
func ConvertStreamChunk(msg *schema.Message, modelName string, id string) *StreamChunk {
	chunk := &StreamChunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   modelName,
		Choices: []Choice{
			{
				Index: 0,
				Delta: &Message{
					Role:    RoleAssistant,
					Content: msg.Content,
				},
			},
		},
	}

	// Extract usage if present (typically in the last chunk)
	if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
		chunk.Usage = &Usage{
			PromptTokens:     msg.ResponseMeta.Usage.PromptTokens,
			CompletionTokens: msg.ResponseMeta.Usage.CompletionTokens,
			TotalTokens:      msg.ResponseMeta.Usage.TotalTokens,
		}
	}

	return chunk
}

// ExtractUsage extracts token usage from Eino message.
func ExtractUsage(msg *schema.Message) Usage {
	if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
		return Usage{
			PromptTokens:     msg.ResponseMeta.Usage.PromptTokens,
			CompletionTokens: msg.ResponseMeta.Usage.CompletionTokens,
			TotalTokens:      msg.ResponseMeta.Usage.TotalTokens,
		}
	}
	return Usage{}
}
```

**Step 3: Create tests**

```go
// pkg/ai/common_test.go
package ai

import (
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestGenerateID(t *testing.T) {
	id := GenerateID()
	if len(id) == 0 {
		t.Fatal("GenerateID returned empty string")
	}
	if id[:8] != "chatcmpl-" {
		t.Fatalf("GenerateID has wrong prefix: %s", id[:8])
	}
}

func TestConvertMessages(t *testing.T) {
	msgs := []Message{
		{Role: RoleSystem, Content: "You are helpful"},
		{Role: RoleUser, Content: "Hello"},
	}

	converted := ConvertMessages(msgs)

	if len(converted) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(converted))
	}
	if converted[0].Role != schema.System {
		t.Errorf("Expected System role, got %v", converted[0].Role)
	}
}

func TestExtractUsage(t *testing.T) {
	msg := &schema.Message{
		ResponseMeta: &schema.ResponseMeta{
			Usage: &schema.Usage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		},
	}

	usage := ExtractUsage(msg)
	if usage.TotalTokens != 30 {
		t.Errorf("Expected 30 tokens, got %d", usage.TotalTokens)
	}
}
```

**Step 4: Remove duplicated code from providers**

In each provider (`openai/provider.go`, `claude/provider.go`, `gemini/provider.go`, `qwen/provider.go`):
- Remove local `generateID`, `convertMessages`, `convertResponse`, `convertStreamChunk`, `extractUsage` functions
- Update calls to use `ai.GenerateID()`, `ai.ConvertMessages()`, etc.

**Step 5: Verify and commit**

```bash
go build ./pkg/ai/...
go test ./pkg/ai/...

git add pkg/ai/common.go pkg/ai/common_test.go pkg/ai/providers/
git commit -m "refactor(ai): extract common utilities to reduce code duplication

- Move GenerateID, ConvertMessages, ConvertResponse to pkg/ai/common.go
- Remove duplicate functions from all 4 providers
- Add unit tests for common utilities
"
```

---

### Task 2: Fix Quota Leak with Defer Pattern

**Files:**
- Modify: `internal/apiserver/biz/chat/chat.go`

**Step 1: Identify the actual quota leak**

The real issue is at **line 84-86** (Chat method) and **line 160-162** (ChatStream method):

```go
// Reserve TPD quota atomically before calling provider
reservedTokens, err := b.quota.ReserveTPD(ctx, uid, req.MaxTokens)
if err != nil {
    return nil, err  // ❌ LEAK: No quota reserved yet, but if ReserveTPD succeeded...
}
```

Actually, `ReserveTPD` returns error *before* reserving, so there's no leak here. **The real leak is:**
- If `GetByModel` fails (line 90, 166) - no quota release
- If `provider.Chat` fails but we return early (line 96, 172) - quota IS released (correct)
- **Missing**: panic or unexpected early return paths

**Step 2: Implement defensive defer pattern**

Use a defer to ensure quota is always released unless explicitly marked as consumed:

```go
// In Chat() method - internal/apiserver/biz/chat/chat.go:84

// Reserve TPD quota atomically
reservedTokens, err := b.quota.ReserveTPD(ctx, uid, req.MaxTokens)
if err != nil {
    return nil, err
}

// Ensure quota is released if not consumed
quotaConsumed := false
defer func() {
    if !quotaConsumed && reservedTokens > 0 {
        // Release in background to avoid blocking
        go func() {
            ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
            defer cancel()
            if err := b.quota.AdjustTPD(ctx, uid, 0, reservedTokens); err != nil {
                log.C(ctx).Errorw("Failed to release reserved quota",
                    "uid", uid, "reserved", reservedTokens, "err", err)
            }
        }()
    }
}()

// Get provider for the model
provider, ok := b.registry.GetByModel(req.Model)
if !ok {
    return nil, errno.ErrAIModelNotFound
    // defer will automatically release quota
}

// Call provider
resp, err := provider.Chat(ctx, req)
if err != nil {
    return nil, errno.ErrAIProviderError.WithMessage("chat failed: %v", err)
    // defer will automatically release quota
}

// Mark quota as consumed (will be adjusted with actual usage below)
quotaConsumed = true

// Adjust TPD quota with actual usage (background)
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
    defer cancel()
    if err := b.quota.AdjustTPD(ctx, uid, resp.Usage.TotalTokens, reservedTokens); err != nil {
        log.C(ctx).Errorw("Failed to adjust TPD quota",
            "uid", uid, "actual", resp.Usage.TotalTokens,
            "reserved", reservedTokens, "err", err)
    }
}()
```

Apply the same pattern to `ChatStream()` method.

**Step 3: Test the fix**

```bash
# Test quota is released when model not found
curl -X POST /api/v1/chat/stream -d '{"model":"nonexistent","messages":[...]}'
# Check Redis TPD quota - should be released immediately

# Test quota is adjusted after successful call
curl -X POST /api/v1/chat/stream -d '{"model":"gpt-4","messages":[...]}'
# Check Redis TPD quota - should reflect actual usage
```

**Step 4: Commit**

```bash
git add internal/apiserver/biz/chat/chat.go
git commit -m "fix(ai): prevent quota leak with defer release pattern

- Add defer to ensure reserved quota is always released on error paths
- Use quotaConsumed flag to prevent double-release
- Apply to both Chat() and ChatStream() methods
- Fixes potential quota leak when provider initialization fails
"
```

---

### Task 3: Fix Message Duplication in History

**Files:**
- Modify: `internal/apiserver/biz/chat/chat.go`

**Step 1: Understand the actual problem**

After code analysis:
- Code already captures `newMessages` at line 74/150 **before** calling `loadAndMergeHistory`
- `loadAndMergeHistory` creates a **new slice**, so `newMessages` is not modified
- **The real issue**: If frontend sends `[OldMsg1, OldMsg2, NewMsg]`, then `newMessages` contains all 3 messages

When saving to session, we iterate all `newMessages` and save user messages, causing duplicates.

**Scenario that triggers the bug:**
```
DB has: [user1, asst1, user2, asst2]
Frontend sends: [user1, asst1, user2, asst2, user3]  (continuation)
Result: saves user1, user2, user3 -> user1 and user2 are duplicated!
```

**Step 2: Fix in `loadAndMergeHistory`**

The best place to fix this is in `loadAndMergeHistory` itself, by filtering new messages when we have history:

```go
// loadAndMergeHistory loads session history and merges with new messages.
// Returns merged messages with sliding window applied.
func (b *chatBiz) loadAndMergeHistory(ctx context.Context, sessionID string, newMessages []ai.Message) ([]ai.Message, error) {
    if sessionID == "" {
        return newMessages, nil
    }

    // Get limit from config
    limit := facade.Config.AI.Session.MaxMessages
    if limit <= 0 {
        limit = 50 // default
    }

    // Load history messages
    history, err := b.ds.AiMessage().ListBySessionID(ctx, sessionID, limit)
    if err != nil {
        log.C(ctx).Warnw("Failed to load message history", "session_id", sessionID, "err", err)
        return newMessages, nil // Continue without history on error
    }

    if len(history) == 0 {
        return newMessages, nil
    }

    // Convert DB messages to ai.Message
    messages := make([]ai.Message, 0, len(history)+len(newMessages))
    for _, m := range history {
        messages = append(messages, ai.Message{
            Role:    m.Role,
            Content: m.Content,
        })
    }

    // FIX: When we have history, only append the NEW user message(s)
    // Find the last user message in newMessages
    if len(newMessages) > 0 {
        lastUserMsgIdx := -1
        for i := len(newMessages) - 1; i >= 0; i-- {
            if newMessages[i].Role == ai.RoleUser {
                lastUserMsgIdx = i
                break
            }
        }

        if lastUserMsgIdx >= 0 {
            // Only include messages from the last user message onwards
            messages = append(messages, newMessages[lastUserMsgIdx:]...)
        } else {
            // No user message in newMessages, append as-is (edge case)
            messages = append(messages, newMessages...)
        }
    }

    // Apply sliding window if configured
    contextWindow := facade.Config.AI.Session.ContextWindow
    if contextWindow > 0 && len(messages) > contextWindow {
        // Keep system message if present, then last N-1 messages
        var result []ai.Message
        if len(messages) > 0 && messages[0].Role == ai.RoleSystem {
            result = append(result, messages[0])
            messages = messages[1:]
            contextWindow--
        }
        if len(messages) > contextWindow {
            messages = messages[len(messages)-contextWindow:]
        }
        result = append(result, messages...)
        messages = result
    }

    return messages, nil
}
```

**Step 3: Update callers**

The callers (`Chat` and `ChatStream`) don't need changes. They already use `newMessages` correctly:

```go
// Already correct - no changes needed
newMessages := req.Messages
messages, err := b.loadAndMergeHistory(ctx, req.SessionID, req.Messages)
req.Messages = messages

// Later - correctly saves only new messages
b.saveToSession(ctx, uid, req.SessionID, newMessages, resp)
```

**Step 4: Commit**

```bash
git add internal/apiserver/biz/chat/chat.go
git commit -m "fix(ai): prevent duplicate messages when session has history

- Filter newMessages in loadAndMergeHistory to only include last user message
- Prevents saving duplicate user messages when frontend sends full conversation
- Maintains conversation continuity while avoiding database duplicates
"
```

---

### Task 4: Decide on Provider Constructor Context Parameter

**Files:**
- Modify: `pkg/ai/providers/gemini/provider.go` (if we decide to remove ctx)
- OR: Modify all other providers to accept ctx (if we decide to keep it)
- Modify: `internal/apiserver/biz/chat/registry.go` (update initialization)

**Design Discussion: Should `New()` accept context?**

This is a **non-trivial design decision** with trade-offs. Let's analyze:

#### Option A: Remove context from Gemini.New() → Make all providers consistent (no ctx)

**Pros:**
- **Consistency**: All 4 providers have the same signature `New(cfg *Config) (*Provider, error)`
- **Simpler initialization**: Callers don't need to manage context during startup
- **No startup cancellation**: Provider initialization is typically fast and non-cancellable anyway

**Cons:**
- **Gemini client requires context**: We must use `context.Background()` internally
- **Lose timeout control**: Can't set a timeout for provider initialization
- **Less flexible**: If initialization becomes slow in the future, we can't add timeout easily

**Current Gemini implementation:**
```go
func New(ctx context.Context, cfg *Config) (*Provider, error) {
    genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{...})
    // ...
}
```

**Proposed change:**
```go
func New(cfg *Config) (*Provider, error) {
    ctx := context.Background()  // No timeout/cancellation
    genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{...})
    // ...
}
```

#### Option B: Add context to ALL providers → Keep Gemini's design

**Pros:**
- **Future-proof**: If any provider needs slow initialization (e.g., warm-up calls), we're ready
- **Timeout control**: Can pass `context.WithTimeout()` during startup
- **Idiomatic Go**: Constructor that might do I/O often accepts context
- **Testability**: Tests can use `context.WithTimeout()` to verify init behavior

**Cons:**
- **Inconsistency with current code**: Need to update OpenAI, Claude, Qwen
- **Caller complexity**: All initialization sites need to pass context
- **Questionable value**: Provider init is usually just creating a client, not doing I/O

#### Option C: Hybrid approach - Use context.Background() only for Gemini

**Keep current code** - Gemini has ctx, others don't.

**Pros:**
- **Minimal changes**: No code modification needed
- **Honest about requirements**: Gemini's client actually needs ctx, others don't

**Cons:**
- **Inconsistent API**: Different providers have different signatures
- **Documentation burden**: Need to document why Gemini is different
- **Registry complexity**: Need to handle both signatures in initialization code

---

#### **Recommendation: Option A - Remove context from Gemini**

**Rationale:**
1. **Provider initialization should be fast** - It's just creating a client object, not making API calls
2. **Consistency is valuable** - Having the same signature across all providers simplifies the registry
3. **Gemini's NewClient doesn't actually block** - Looking at the code, it's just client initialization, not an API call
4. **If initialization becomes slow**, we can add it back with a better design (e.g., `Initialize(ctx)` method)

**Evidence from code:**
```go
// All providers use context.Background() in their Chat/ChatStream methods
// This suggests initialization doesn't need context
client, err := openai.NewChatModel(context.Background(), ...)
```

**If later we need context for init**, we can introduce:
```go
func (p *Provider) Initialize(ctx context.Context) error {
    // Expensive initialization, warm-up calls, etc.
}
```

**Step 1: Update Gemini provider**

```go
// pkg/ai/providers/gemini/provider.go:30

// New creates a new Gemini provider
func New(cfg *Config) (*Provider, error) {
    ctx := context.Background()
    genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey:  cfg.APIKey,
        Backend: genai.BackendGeminiAPI,
    })
    if err != nil {
        return nil, err
    }

    client, err := gemini.NewChatModel(ctx, &gemini.Config{
        Client: genaiClient,
        Model:  cfg.Models[0].ID,
    })
    if err != nil {
        return nil, err
    }

    return &Provider{
        config: cfg,
        client: client,
    }, nil
}
```

**Step 2: Update registry initialization**

Find and update caller:
```go
// Before
geminiProv, err := gemini.New(ctx, geminiCfg)

// After
geminiProv, err := gemini.New(geminiCfg)
```

**Step 3: Commit**

```bash
git add pkg/ai/providers/gemini/provider.go internal/apiserver/
git commit -m "refactor(ai): standardize provider constructor signatures

- Remove context parameter from gemini.New to match other providers
- Provider initialization is fast and doesn't need cancellation
- Simplifies registry initialization with consistent API
"
```

---

**Decision: Option A selected** - Remove context from Gemini.New()

---

### Task 5: Fix Claude Name Hardcoding

**Files:**
- Modify: `pkg/ai/providers/claude/provider.go`

**Step 1: Update Name() method**

```go
func (p *Provider) Name() string {
    if p.config.Name != "" {
        return p.config.Name
    }
    return "claude"
}
```

**Step 2: Commit**

```bash
git add pkg/ai/providers/claude/provider.go
git commit -m "fix: support custom name in Claude provider"
```

---
