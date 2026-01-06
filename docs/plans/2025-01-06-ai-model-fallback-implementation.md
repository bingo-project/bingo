# AI Model Fallback Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现模型降级机制，当主模型不可用时自动切换到备用模型

**Architecture:** 在 `internal/pkg/ai/` 新增 `FallbackSelector` 组件，`ChatBiz` 注入使用。降级基于数据库 `ai_model` 表的 `sort` 字段排序，`allow_fallback` 字段控制是否参与降级。

**Tech Stack:** Go 1.24+, GORM, Gin

---

## Task 1: Database Migration - Add allow_fallback Field

**Files:**
- Modify: `internal/pkg/database/migration/2025_12_29_100001_create_ai_model_table.go`
- Modify: `internal/pkg/model/ai_model.go`

**Step 1: Modify migration file to add allow_fallback field**

Edit the `CreateAIModelTable` struct, add `AllowFallback` field before `CreatedAt`:

```go
type CreateAIModelTable struct {
    ID           uint64    `gorm:"primaryKey"`
    ProviderName string    `gorm:"type:varchar(32);index:idx_provider_name;not null"`
    Model        string    `gorm:"type:varchar(64);uniqueIndex:uk_model;not null"`
    DisplayName  string    `gorm:"type:varchar(64)"`
    MaxTokens    int       `gorm:"type:int;not null;default:4096"`
    InputPrice   float64   `gorm:"type:decimal(10,6);not null;default:0"`
    OutputPrice  float64   `gorm:"type:decimal(10,6);not null;default:0"`
    Status       string    `gorm:"type:varchar(16);not null;default:active"`
    IsDefault    bool      `gorm:"type:tinyint(1);not null;default:0"`
    Sort         int       `gorm:"type:int;not null;default:0"`
    AllowFallback bool     `gorm:"type:tinyint(1);not null;default:1;comment:是否允许作为降级目标"`
    CreatedAt    time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
    UpdatedAt    time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}
```

**Step 2: Modify model struct to add allow_fallback field**

Edit `internal/pkg/model/ai_model.go`, add `AllowFallback` field to `AiModelM` struct:

```go
type AiModelM struct {
    ID           uint    `gorm:"primaryKey" json:"id"`
    ProviderName string  `gorm:"column:provider_name;type:varchar(32);index:idx_provider_name;not null" json:"providerName"`
    Model        string  `gorm:"column:model;type:varchar(64);uniqueIndex:uk_model;not null" json:"model"`
    DisplayName  string  `gorm:"column:display_name;type:varchar(64)" json:"displayName"`
    MaxTokens    int     `gorm:"column:max_tokens;type:int;not null;default:4096" json:"maxTokens"`
    InputPrice   float64 `gorm:"column:input_price;type:decimal(10,6);not null;default:0" json:"inputPrice"`
    OutputPrice  float64 `gorm:"column:output_price;type:decimal(10,6);not null;default:0" json:"outputPrice"`
    Status       string  `gorm:"column:status;type:varchar(16);not null;default:active" json:"status"`
    IsDefault    bool    `gorm:"column:is_default;type:tinyint(1);not null;default:0" json:"isDefault"`
    Sort         int     `gorm:"column:sort;type:int;not null;default:0" json:"sort"`
    AllowFallback bool   `gorm:"column:allow_fallback;type:tinyint(1);not null;default:1" json:"allowFallback"`

    CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"createdAt"`
    UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}
```

**Step 3: Verify migration**

Run: `bingo migrate reset && bingo migrate up`
Expected: No errors, table created with new field

**Step 4: Commit**

```bash
git add internal/pkg/database/migration/2025_12_29_100001_create_ai_model_table.go internal/pkg/model/ai_model.go
git commit -m "feat(ai): add allow_fallback field to ai_model table"
```

---

## Task 2: Add Error Code for All Models Failed

**Files:**
- Modify: `internal/pkg/errno/ai.go`

**Step 1: Add new error code**

Add to `internal/pkg/errno/ai.go`:

```go
// ErrAIAllModelsFailed indicates all models (including fallback) failed
var ErrAIAllModelsFailed = &errorsx.ErrorX{
    Code:    http.StatusServiceUnavailable,
    Reason:  "ServiceUnavailable.AllModelsFailed",
    Message: "AI service is temporarily unavailable, please try again later.",
}
```

**Step 2: Verify code compiles**

Run: `go build ./internal/pkg/errno/...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/pkg/errno/ai.go
git commit -m "feat(errno): add ErrAIAllModelsFailed error code"
```

---

## Task 3: Create FallbackSelector Component

**Files:**
- Create: `internal/pkg/ai/fallback.go`
- Create: `internal/pkg/ai/fallback_test.go`

**Step 1: Write the failing test first (TDD)**

Create `internal/pkg/ai/fallback_test.go`:

```go
// ABOUTME: Tests for AI model fallback selection.
// ABOUTME: Verifies fallback model selection logic.

package ai

import (
    "context"

    "github.com/bingo-project/bingo/internal/pkg/model"
    "github.com/bingo-project/bingo/internal/pkg/store"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// mockModelStore is a minimal mock for testing
type mockModelStore struct {
    models []*model.AiModelM
    err    error
}

func (m *mockModelStore) ListActive(context.Context) ([]*model.AiModelM, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.models, nil
}

// Implement other required methods as no-ops
func (m *mockModelStore) Create(ctx context.Context, obj *model.AiModelM) error { return nil }
func (m *mockModelStore) Update(ctx context.Context, obj *model.AiModelM, fields ...string) error { return nil }
func (m *mockModelStore) Delete(ctx context.Context, opts *store.WhereOptions) error { return nil }
func (m *mockModelStore) Get(ctx context.Context, opts *store.WhereOptions) (*model.AiModelM, error) { return nil, nil }
func (m *mockModelStore) List(ctx context.Context, opts *store.WhereOptions) (int64, []*model.AiModelM, error) { return 0, nil, nil }
func (m *mockModelStore) GetByModel(ctx context.Context, modelID string) (*model.AiModelM, error) { return nil, nil }
func (m *mockModelStore) ListByProvider(ctx context.Context, providerName string) ([]*model.AiModelM, error) { return nil, nil }
func (m *mockModelStore) GetDefault(ctx context.Context) (*model.AiModelM, error) { return nil, nil }
func (m *mockModelStore) FirstOrCreate(ctx context.Context, where *model.AiModelM, obj *model.AiModelM) error { return nil }

func TestFallbackSelector_SelectFallback_Success(t *testing.T) {
    registry := NewRegistry()

    // Register test providers
    registry.Register(&mockProvider{name: "openai", models: []ModelInfo{
        {ID: "gpt-4o", Provider: "openai"},
        {ID: "gpt-3.5-turbo", Provider: "openai"},
    }})

    store := &mockModelStore{
        models: []*model.AiModelM{
            {Model: "gpt-4o", AllowFallback: true, Sort: 1, Status: model.AiModelStatusActive},
            {Model: "gpt-3.5-turbo", AllowFallback: true, Sort: 2, Status: model.AiModelStatusActive},
            {Model: "claude-3-5-sonnet", AllowFallback: false, Sort: 3, Status: model.AiModelStatusActive},
        },
    }

    selector := NewFallbackSelector(store, registry)

    fallback := selector.SelectFallback(context.Background(), "gpt-4o")

    assert.Equal(t, "gpt-3.5-turbo", fallback, "should return next model by sort order")
}

func TestFallbackSelector_SelectFallback_NoFallbackAllowed(t *testing.T) {
    registry := NewRegistry()
    registry.Register(&mockProvider{name: "openai", models: []ModelInfo{
        {ID: "gpt-4o", Provider: "openai"},
    }})

    store := &mockModelStore{
        models: []*model.AiModelM{
            {Model: "gpt-4o", AllowFallback: false, Sort: 1, Status: model.AiModelStatusActive},
        },
    }

    selector := NewFallbackSelector(store, registry)

    fallback := selector.SelectFallback(context.Background(), "gpt-4o")

    assert.Equal(t, "", fallback, "should return empty when no fallback allowed")
}

func TestFallbackSelector_SelectFallback_NoModelsRegistered(t *testing.T) {
    registry := NewRegistry()

    store := &mockModelStore{
        models: []*model.AiModelM{
            {Model: "gpt-4o", AllowFallback: true, Sort: 1, Status: model.AiModelStatusActive},
        },
    }

    selector := NewFallbackSelector(store, registry)

    fallback := selector.SelectFallback(context.Background(), "unknown-model")

    assert.Equal(t, "", fallback, "should return empty when model not found")
}

func TestFallbackSelector_SelectFallback_StoreError(t *testing.T) {
    registry := NewRegistry()

    store := &mockModelStore{
        err: assert.AnError,
    }

    selector := NewFallbackSelector(store, registry)

    fallback := selector.SelectFallback(context.Background(), "gpt-4o")

    assert.Equal(t, "", fallback, "should return empty on store error")
}

// mockProvider is a minimal mock Provider
type mockProvider struct {
    name   string
    models []ModelInfo
}

func (m *mockProvider) Name() string        { return m.name }
func (m *mockProvider) Models() []ModelInfo  { return m.models }
func (m *mockProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    return nil, nil
}
func (m *mockProvider) ChatStream(ctx context.Context, req *ChatRequest) (*ChatStream, error) {
    return nil, nil
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/pkg/ai/fallback_test.go -v`
Expected: FAIL with "undefined: NewFallbackSelector"

**Step 3: Implement FallbackSelector**

Create `internal/pkg/ai/fallback.go`:

```go
// ABOUTME: AI model fallback selection logic.
// ABOUTME: Handles model degradation when primary model is unavailable.

package ai

import (
    "context"

    "github.com/bingo-project/bingo/internal/pkg/log"
    "github.com/bingo-project/bingo/internal/pkg/store"
)

// FallbackSelector handles model fallback selection
type FallbackSelector struct {
    store    store.AiModelStore
    registry *Registry
}

// NewFallbackSelector creates a new FallbackSelector
func NewFallbackSelector(store store.AiModelStore, registry *Registry) *FallbackSelector {
    return &FallbackSelector{
        store:    store,
        registry: registry,
    }
}

// SelectFallback returns the next available model for fallback.
// Returns empty string if no fallback model is available.
func (s *FallbackSelector) SelectFallback(ctx context.Context, originalModel string) string {
    models, err := s.store.ListActive(ctx)
    if err != nil {
        log.C(ctx).Warnw("Failed to list models for fallback", "err", err)
        return ""
    }

    for _, m := range models {
        // Skip original model
        if m.Model == originalModel {
            continue
        }
        // Check if fallback is allowed
        if !m.AllowFallback {
            continue
        }
        // Check if model is registered in Registry
        if _, ok := s.registry.GetByModel(m.Model); ok {
            log.C(ctx).Infow("AI model fallback selected",
                "original", originalModel,
                "fallback", m.Model)
            return m.Model
        }
    }

    return ""
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/pkg/ai/... -run TestFallbackSelector -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/pkg/ai/fallback.go internal/pkg/ai/fallback_test.go
git commit -m "feat(ai): add FallbackSelector for model degradation"
```

---

## Task 4: Export FallbackSelector from ai.go

**Files:**
- Modify: `internal/pkg/ai/ai.go`

**Step 1: Export FallbackSelector type**

Edit `internal/pkg/ai/ai.go` to re-export the type (if needed for cross-package usage):

Check if `ai.go` needs modification. The `FallbackSelector` is already in `internal/pkg/ai` package, so it should be directly accessible by `internal/apiserver/biz/chat`. No changes needed if the import path is correct.

Verify `ChatBiz` can import: `github.com/bingo-project/bingo/internal/pkg/ai`

**Step 2: Commit (if changes made)**

If no changes needed, skip this step.

---

## Task 5: Fix retry.go - Remove insufficient_quota from retriable errors

**Files:**
- Modify: `pkg/ai/retry.go`

**Step 1: Remove insufficient_quota from retriable list**

Edit `pkg/ai/retry.go`, remove `"insufficient_quota"` from `retriableMessages`:

```go
func isRetriable(err error) bool {
    if err == nil {
        return false
    }

    errMsg := strings.ToLower(err.Error())
    retriableMessages := []string{
        "429",                   // Too Many Requests
        "503",                   // Service Unavailable
        "502",                   // Bad Gateway
        "504",                   // Gateway Timeout
        "timeout",               // General timeout
        "deadline exceeded",     // Context deadline exceeded
        "connection refused",    // Network issue
        "connection reset",      // Network issue
        "request_timeout_error", // Specific provider error
        "rate_limit_reached",    // Specific provider error
        // "insufficient_quota",  // REMOVED: quota不足不应重试
        "overloaded",            // Specific provider error (e.g. Claude)
    }

    for _, m := range retriableMessages {
        if strings.Contains(errMsg, m) {
            return true
        }
    }

    return false
}
```

**Step 2: Verify code compiles**

Run: `go build ./pkg/ai/...`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/ai/retry.go
git commit -m "fix(ai): remove insufficient_quota from retriable errors"
```

---

## Task 6: Integrate FallbackSelector into ChatBiz

**Files:**
- Modify: `internal/apiserver/biz/chat/chat.go`

**Step 1: Add fallback field to chatBiz struct**

Add `fallback *ai.FallbackSelector` field to `chatBiz`:

```go
type chatBiz struct {
    ds       store.IStore
    registry *ai.Registry
    quota    *quotaChecker
    fallback *ai.FallbackSelector  // NEW
}
```

**Step 2: Initialize fallback in New() constructor**

```go
func New(ds store.IStore, registry *ai.Registry) *chatBiz {
    return &chatBiz{
        ds:       ds,
        registry: registry,
        quota:    newQuotaChecker(ds),
        fallback: ai.NewFallbackSelector(ds.AiModel(), registry),  // NEW
    }
}
```

**Step 3: Add helper method getProviderWithFallback**

Add after the `Chat` method (around line 146):

```go
// getProviderWithFallback gets provider with fallback support.
// Returns (provider, actualModelUsed, error).
func (b *chatBiz) getProviderWithFallback(ctx context.Context, model string) (ai.Provider, string, error) {
    // First attempt: get provider for requested model
    provider, ok := b.registry.GetByModel(model)
    if ok {
        return provider, model, nil
    }

    // Fallback attempt: model not registered, try fallback
    fallbackModel := b.fallback.SelectFallback(ctx, model)
    if fallbackModel == "" {
        return nil, "", errno.ErrAIModelNotFound
    }

    provider, ok = b.registry.GetByModel(fallbackModel)
    if !ok {
        return nil, "", errno.ErrAIAllModelsFailed
    }

    return provider, fallbackModel, nil
}

// isRetriableProviderError checks if error should trigger fallback retry.
func (b *chatBiz) isRetriableProviderError(err error) bool {
    if err == nil {
        return false
    }

    errMsg := strings.ToLower(err.Error())
    retriable := []string{"429", "503", "502", "504", "timeout", "overloaded"}
    for _, r := range retriable {
        if strings.Contains(errMsg, r) {
            return true
        }
    }
    return false
}
```

**Step 4: Modify Chat method to use getProviderWithFallback**

Replace the provider lookup section in `Chat` method (around line 108-120):

Find:
```go
// Get provider for the model
provider, ok := b.registry.GetByModel(req.Model)
if !ok {
    return nil, errno.ErrAIModelNotFound
}

// Call provider
resp, err := provider.Chat(ctx, req)
```

Replace with:
```go
// Get provider with fallback
provider, modelUsed, err := b.getProviderWithFallback(ctx, req.Model)
if err != nil {
    return nil, err
}
req.Model = modelUsed

// Call provider
resp, err := provider.Chat(ctx, req)
if err != nil {
    // Check if error is retriable and we haven't tried fallback yet
    if b.isRetriableProviderError(err) {
        fallbackModel := b.fallback.SelectFallback(ctx, modelUsed)
        if fallbackModel != "" {
            if provider2, ok := b.registry.GetByModel(fallbackModel); ok {
                log.C(ctx).Infow("AI provider error, using fallback",
                    "model", modelUsed, "fallback", fallbackModel, "err", err)
                req.Model = fallbackModel
                resp, err = provider2.Chat(ctx, req)
                if err == nil {
                    return resp, nil
                }
            }
        }
    }
    return nil, errno.ErrAIProviderError.WithMessage("chat failed: %v", err)
}
```

**Step 5: Modify ChatStream method similarly**

Apply the same changes to `ChatStream` method (around line 196-207).

**Step 6: Verify code compiles**

Run: `go build ./internal/apiserver/biz/chat/...`
Expected: No errors

**Step 7: Run tests**

Run: `go test ./internal/apiserver/biz/chat/... -v`
Expected: Existing tests still pass

**Step 8: Commit**

```bash
git add internal/apiserver/biz/chat/chat.go
git commit -m "feat(chat): integrate FallbackSelector for model degradation"
```

---

## Task 7: Improve resolveModel to use database default

**Files:**
- Modify: `internal/apiserver/biz/chat/chat.go`

**Step 1: Update resolveModel method**

Replace the `resolveModel` method (around line 353-375):

```go
// resolveModel resolves the model to use based on priority:
// Request specified > Session preference > Database default > Config default > First available
func (b *chatBiz) resolveModel(ctx context.Context, reqModel, sessionID string) string {
    // 1. Request specified
    if reqModel != "" {
        return reqModel
    }

    // 2. Session preference (if session has a model set)
    if sessionID != "" {
        session, err := b.ds.AiSession().GetBySessionID(ctx, sessionID)
        if err == nil && session.Model != "" {
            return session.Model
        }
    }

    // 3. Database default model (is_default=true)
    defaultModel, err := b.ds.AiModel().GetDefault(ctx)
    if err == nil && defaultModel != nil {
        return defaultModel.Model
    }

    // 4. Config file fallback
    if facade.Config.AI.DefaultModel != "" {
        return facade.Config.AI.DefaultModel
    }

    // 5. First available model by sort order
    models, err := b.ds.AiModel().ListActive(ctx)
    if err == nil && len(models) > 0 {
        return models[0].Model
    }

    return "" // Empty string - let caller handle error
}
```

**Step 2: Update Chat/ChatStream to handle empty model**

After `resolveModel` call in both `Chat` and `ChatStream`, add check:

```go
// Resolve model
req.Model = b.resolveModel(ctx, req.Model, req.SessionID)
if req.Model == "" {
    return nil, errno.ErrAIModelNotFound
}
```

**Step 3: Verify code compiles**

Run: `go build ./internal/apiserver/biz/chat/...`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/apiserver/biz/chat/chat.go
git commit -m "feat(chat): improve resolveModel with database default"
```

---

## Task 8: End-to-End Testing

**Files:**
- None (manual testing)

**Step 1: Build all services**

Run: `make build`
Expected: All binaries compiled successfully

**Step 2: Run database migration**

Run: `bingo migrate reset && bingo migrate up && bingo db seed`
Expected: Database created with seed data

**Step 3: Start apiserver**

Run: `./bin/apiserver`
Expected: Server starts without errors

**Step 4: Test normal chat (no fallback)**

Run:
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```
Expected: Normal response

**Step 5: Test fallback (model not registered)**

Run with non-existent model:
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "non-existent-model",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```
Expected: Falls back to next available model, check logs for fallback message

**Step 6: Check logs**

Expected log output:
```
INFO AI model fallback selected original=non-existent-model fallback=gpt-3.5-turbo
```

**Step 7: Test all models failed scenario**

Disable all models in database, then retry request.

Expected: HTTP 503 with message "AI service is temporarily unavailable, please try again later."

---

## Task 9: Run Lint and Final Verification

**Step 1: Run linter**

Run: `make lint`
Expected: No errors

**Step 2: Run all tests**

Run: `make test`
Expected: All tests pass

**Step 3: Verify Swagger docs (if API changed)**

Run: `make swag` (if any API structures changed)
Expected: Swagger docs updated

**Step 4: Final build**

Run: `make build`
Expected: Clean build

**Step 5: Final commit**

```bash
git add -A
git commit -m "chore: final verification complete"
```

---

## Summary of File Changes

| File | Type | Description |
|------|------|-------------|
| `internal/pkg/database/migration/..._create_ai_model_table.go` | Modify | Add AllowFallback field |
| `internal/pkg/model/ai_model.go` | Modify | Add AllowFallback field |
| `internal/pkg/errno/ai.go` | Modify | Add ErrAIAllModelsFailed |
| `internal/pkg/ai/fallback.go` | Create | FallbackSelector implementation |
| `internal/pkg/ai/fallback_test.go` | Create | Unit tests |
| `pkg/ai/retry.go` | Modify | Remove insufficient_quota |
| `internal/apiserver/biz/chat/chat.go` | Modify | Integrate fallback |

## Verification Checklist

- [ ] All migrations applied
- [ ] All unit tests pass
- [ ] Linter passes
- [ ] E2E test passes (normal chat)
- [ ] E2E test passes (fallback triggered)
- [ ] E2E test passes (all models failed)
- [ ] Logs show fallback events
