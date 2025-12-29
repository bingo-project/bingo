# AI 对话功能实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 Bingo 添加通用 AI 对话能力，支持多 Provider、会话管理、流式输出、按用户配额限流。

**Architecture:** 三层架构 (Handler → Biz → Store)，pkg/ai 提供通用 AI 能力，内部用 Eino 框架但通过 Interface 解耦。OpenAI 兼容 API + 扩展字段。

**Tech Stack:** Go 1.24+, Gin, GORM, Redis, Eino (CloudWeGo), ulule/limiter

**Design Doc:** [2025-12-29-ai-chat-design.md](./2025-12-29-ai-chat-design.md)

**Conventions:** [CONVENTIONS.md](../guides/CONVENTIONS.md)

---

## Phase 1: 基础设施 (Database + Config + Errors)

### Task 1.1: 创建 AI 相关数据库迁移

**Files:**
- Create: `internal/pkg/database/migration/2025_12_29_100000_create_ai_provider_table.go`
- Create: `internal/pkg/database/migration/2025_12_29_100001_create_ai_model_table.go`
- Create: `internal/pkg/database/migration/2025_12_29_100002_create_ai_quota_tier_table.go`
- Create: `internal/pkg/database/migration/2025_12_29_100003_create_ai_user_quota_table.go`
- Create: `internal/pkg/database/migration/2025_12_29_100004_create_ai_session_table.go`
- Create: `internal/pkg/database/migration/2025_12_29_100005_create_ai_message_table.go`

**Step 1: 创建 ai_provider 表迁移**

```go
// internal/pkg/database/migration/2025_12_29_100000_create_ai_provider_table.go

// ABOUTME: Database migration for ai_provider table.
// ABOUTME: Creates table for AI provider configuration.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateAiProviderTable struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"type:varchar(32);uniqueIndex:uk_name;not null"`
	DisplayName string    `gorm:"type:varchar(64)"`
	Status      string    `gorm:"type:varchar(16);not null;default:active"`
	Models      string    `gorm:"type:json"`
	IsDefault   bool      `gorm:"type:tinyint(1);not null;default:0"`
	Sort        int       `gorm:"type:int;not null;default:0"`
	CreatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
	UpdatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (CreateAiProviderTable) TableName() string {
	return "ai_provider"
}

func (CreateAiProviderTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateAiProviderTable{})
}

func (CreateAiProviderTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateAiProviderTable{})
}

func init() {
	migrate.Add("2025_12_29_100000_create_ai_provider_table", CreateAiProviderTable{}.Up, CreateAiProviderTable{}.Down)
}
```

**Step 2: 创建其他 5 个迁移文件**

按相同模式创建：
- `ai_model` 表 (provider_name, model, display_name, max_tokens, input_price, output_price, status, is_default, sort)
- `ai_quota_tier` 表 (tier, rpm, tpd)
- `ai_user_quota` 表 (uid, tier, rpm, tpd)
- `ai_session` 表 (session_id, uid, title, model, message_count, total_tokens, status)
- `ai_message` 表 (session_id, role, content, tokens, model)

**Step 3: 运行迁移验证**

```bash
bingo migrate up                    # 执行迁移
bingo migrate rollback              # 验证回滚
bingo migrate up                    # 重新执行
```
Expected: 6 tables created successfully, rollback 无报错

**Step 4: Commit**

```bash
make lint
git add internal/pkg/database/migration/2025_12_29_*.go
git commit -m "feat(ai): add database migrations for AI tables"
```

---

### Task 1.2: 创建 AI Model 定义

**Files:**
- Create: `internal/pkg/model/ai_provider.go`
- Create: `internal/pkg/model/ai_model.go`
- Create: `internal/pkg/model/ai_quota.go`
- Create: `internal/pkg/model/ai_session.go`
- Create: `internal/pkg/model/ai_message.go`

**Step 1: 创建 ai_provider Model**

```go
// internal/pkg/model/ai_provider.go

// ABOUTME: AI provider model definition.
// ABOUTME: Represents AI service providers like OpenAI, DeepSeek.

package model

type AiProviderM struct {
	Base

	Name        string `gorm:"column:name;type:varchar(32);uniqueIndex:uk_name;not null" json:"name"`
	DisplayName string `gorm:"column:display_name;type:varchar(64)" json:"displayName"`
	Status      string `gorm:"column:status;type:varchar(16);not null;default:active" json:"status"`
	Models      string `gorm:"column:models;type:json" json:"models"`
	IsDefault   bool   `gorm:"column:is_default;type:tinyint(1);not null;default:0" json:"isDefault"`
	Sort        int    `gorm:"column:sort;type:int;not null;default:0" json:"sort"`
}

func (*AiProviderM) TableName() string {
	return "ai_provider"
}

// Provider status constants
const (
	AiProviderStatusActive   = "active"
	AiProviderStatusDisabled = "disabled"
)
```

**Step 2: 创建其他 4 个 Model 文件**

按项目规范创建，包含 ABOUTME 注释、TableName() 方法、状态常量。

**Step 3: Commit**

```bash
make lint
git add internal/pkg/model/ai_*.go
git commit -m "feat(ai): add AI model definitions"
```

---

### Task 1.3: 创建 AI 错误码

**Files:**
- Create: `internal/pkg/errno/ai.go`

**Step 1: 创建 AI 错误码文件**

```go
// internal/pkg/errno/ai.go

// ABOUTME: AI module error codes.
// ABOUTME: Defines errors for chat, session, quota operations.

package errno

import (
	"net/http"

	"github.com/bingo-project/bingo/pkg/errorsx"
)

var (
	// ErrAIModelNotFound 模型不存在
	ErrAIModelNotFound = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AIModelNotFound",
		Message: "AI model not found.",
	}

	// ErrAIProviderNotConfigured Provider 未配置
	ErrAIProviderNotConfigured = &errorsx.ErrorX{
		Code:    http.StatusServiceUnavailable,
		Reason:  "InternalError.AIProviderNotConfigured",
		Message: "AI provider is not configured.",
	}

	// ErrAIQuotaExceeded 配额超限
	ErrAIQuotaExceeded = &errorsx.ErrorX{
		Code:    http.StatusTooManyRequests,
		Reason:  "ResourceExhausted.AIQuotaExceeded",
		Message: "AI quota exceeded.",
	}

	// ErrAISessionNotFound 会话不存在
	ErrAISessionNotFound = &errorsx.ErrorX{
		Code:    http.StatusNotFound,
		Reason:  "NotFound.AISessionNotFound",
		Message: "AI session not found.",
	}

	// ErrAIStreamError 流式响应错误
	ErrAIStreamError = &errorsx.ErrorX{
		Code:    http.StatusInternalServerError,
		Reason:  "InternalError.AIStreamError",
		Message: "AI stream error.",
	}

	// ErrAIProviderError Provider 返回错误
	ErrAIProviderError = &errorsx.ErrorX{
		Code:    http.StatusBadGateway,
		Reason:  "ExternalError.AIProviderError",
		Message: "AI provider returned an error.",
	}
)
```

**Step 2: Commit**

```bash
make lint
git add internal/pkg/errno/ai.go
git commit -m "feat(ai): add AI error codes"
```

---

### Task 1.4: 添加 AI 配置结构

**Files:**
- Modify: `internal/apiserver/config/config.go`
- Modify: `configs/bingo-apiserver.example.yaml`

**Step 1: 在 config.go 添加 AI 配置结构**

在现有 Config 结构体中添加 AI 字段，并定义相关子结构体。

**Step 2: 在 example.yaml 添加 AI 配置块**

添加 ai 配置块，包含 default_model, credentials, session, quota。

**Step 3: Commit**

```bash
make lint
git add internal/apiserver/config/config.go configs/bingo-apiserver.example.yaml
git commit -m "feat(ai): add AI configuration"
```

---

## Phase 2: pkg/ai 核心模块

### Task 2.1: 创建 pkg/ai 基础结构

**Files:**
- Create: `pkg/ai/errors.go`
- Create: `pkg/ai/message.go`
- Create: `pkg/ai/provider.go`
- Create: `pkg/ai/registry.go`
- Create: `pkg/ai/config.go`

**Step 1: 创建 errors.go**

```go
// pkg/ai/errors.go

// ABOUTME: AI package error definitions.
// ABOUTME: Contains sentinel errors for AI operations.

package ai

import "errors"

var (
	ErrModelNotFound    = errors.New("model not found")
	ErrProviderNotFound = errors.New("provider not found")
	ErrStreamClosed     = errors.New("stream closed")
)
```

**Step 2: 创建 message.go**

定义 Message, ChatRequest, ChatResponse, ChatStream, Usage, Choice 等 OpenAI 兼容结构。

**Step 3: 创建 provider.go**

```go
// pkg/ai/provider.go

// ABOUTME: Provider interface definition.
// ABOUTME: Abstracts AI service providers for multi-provider support.

package ai

import "context"

// Provider defines the interface for AI service providers.
type Provider interface {
	// Name returns the provider identifier (e.g., "openai", "deepseek").
	Name() string

	// Chat performs a non-streaming chat completion.
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream performs a streaming chat completion.
	ChatStream(ctx context.Context, req *ChatRequest) (*ChatStream, error)

	// Models returns the list of models supported by this provider.
	Models() []ModelInfo
}

// ModelInfo contains model metadata.
type ModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}
```

**Step 4: 创建 registry.go**

实现 Registry 结构体，包含 Register, Get, ListModels 方法。

**Step 5: 创建 config.go**

定义 pkg/ai 的配置结构（与 internal/apiserver/config 中的一致）。

**Step 6: 写单元测试**

Create: `pkg/ai/registry_test.go`
测试 Registry 的 Register, Get, ListModels 方法。

**Step 7: Commit**

```bash
make lint
git add pkg/ai/*.go
git commit -m "feat(ai): add pkg/ai core types and interfaces"
```

---

### Task 2.2: 实现 OpenAI Provider

**Files:**
- Create: `pkg/ai/providers/openai/provider.go`
- Create: `pkg/ai/providers/openai/stream.go`
- Create: `pkg/ai/providers/openai/provider_test.go`

**Step 1: 添加 Eino 依赖**

Run: `go get github.com/cloudwego/eino@latest && go get github.com/cloudwego/eino-ext/components/model/openai@latest`

**Step 2: 写单元测试（TDD）**

Create: `pkg/ai/providers/openai/provider_test.go`
先定义测试用例：
- TestProvider_Name
- TestProvider_Chat (mock HTTP)
- TestProvider_ChatStream (mock HTTP)
- TestProvider_Models

Run: `go test ./pkg/ai/providers/openai/... -v`
Expected: 测试失败（实现未完成）

**Step 3: 实现 OpenAI Provider**

使用 Eino 的 openai 组件实现 Provider 接口。支持 OpenAI 及兼容服务（DeepSeek、Moonshot）。

**Step 4: 实现流式响应处理**

封装 Eino 的 StreamReader 为 ChatStream。

**Step 5: 运行测试验证**

Run: `go test ./pkg/ai/providers/openai/... -v`
Expected: 所有测试通过

**Step 6: Commit**

```bash
make lint
git add pkg/ai/providers/openai/*.go go.mod go.sum
git commit -m "feat(ai): implement OpenAI provider with Eino"
```

---

## Phase 3: Store 层

### Task 3.1: 创建 AI Store

**Files:**
- Create: `internal/pkg/store/ai_provider.go`
- Create: `internal/pkg/store/ai_model.go`
- Create: `internal/pkg/store/ai_quota_tier.go`
- Create: `internal/pkg/store/ai_user_quota.go`
- Create: `internal/pkg/store/ai_session.go`
- Create: `internal/pkg/store/ai_message.go`
- Create: `internal/pkg/store/ai_session_test.go`
- Modify: `internal/pkg/store/store.go`

**Step 1: 创建 ai_session.go**

```go
// internal/pkg/store/ai_session.go

// ABOUTME: AI session data access layer.
// ABOUTME: Provides CRUD operations for AI chat sessions.

package store

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type AiSessionStore interface {
	Create(ctx context.Context, obj *model.AiSessionM) error
	Update(ctx context.Context, obj *model.AiSessionM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AiSessionM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AiSessionM, error)

	AiSessionExpansion
}

type AiSessionExpansion interface {
	GetBySessionID(ctx context.Context, sessionID string) (*model.AiSessionM, error)
	ListByUID(ctx context.Context, uid string, status string) ([]*model.AiSessionM, error)
}

type aiSessionStore struct {
	*genericstore.Store[model.AiSessionM]
}

var _ AiSessionStore = (*aiSessionStore)(nil)

func NewAiSessionStore(store *datastore) *aiSessionStore {
	return &aiSessionStore{
		Store: genericstore.NewStore[model.AiSessionM](store, NewLogger()),
	}
}

func (s *aiSessionStore) GetBySessionID(ctx context.Context, sessionID string) (*model.AiSessionM, error) {
	var session model.AiSessionM
	err := s.DB(ctx).Where("session_id = ?", sessionID).First(&session).Error
	return &session, err
}

func (s *aiSessionStore) ListByUID(ctx context.Context, uid string, status string) ([]*model.AiSessionM, error) {
	var sessions []*model.AiSessionM
	db := s.DB(ctx).Where("uid = ?", uid)
	if status != "" {
		db = db.Where("status = ?", status)
	}
	err := db.Order("updated_at DESC").Find(&sessions).Error
	return sessions, err
}
```

**Step 2: 创建其他 5 个 Store 文件**

按相同模式创建，包含 ABOUTME 注释、接口定义、Expansion 接口。

**Step 3: 修改 store.go 注入新 Store**

在 IStore 接口添加方法，在 datastore 实现。

**Step 4: 写 Store 层测试（SQLite）**

Create: `internal/pkg/store/ai_session_test.go`

使用 SQLite 内存数据库测试：
- TestAiSessionStore_Create
- TestAiSessionStore_GetBySessionID
- TestAiSessionStore_ListByUID
- TestAiSessionStore_Update
- TestAiSessionStore_Delete

Run: `go test ./internal/pkg/store/... -run TestAiSession -v`
Expected: 所有测试通过

**Step 5: Commit**

```bash
make lint
git add internal/pkg/store/ai_*.go internal/pkg/store/store.go
git commit -m "feat(ai): add AI store implementations"
```

---

## Phase 4: Biz 层

### Task 4.1: 创建 Chat Biz

**Files:**
- Modify: `internal/pkg/testing/mock/store/store.go` (扩展 AI Store mock)
- Create: `internal/apiserver/biz/chat/chat.go`
- Create: `internal/apiserver/biz/chat/session.go`
- Create: `internal/apiserver/biz/chat/history.go`
- Create: `internal/apiserver/biz/chat/quota.go`
- Create: `internal/apiserver/biz/chat/chat_test.go`
- Modify: `internal/apiserver/biz/biz.go`

**Step 1: 扩展 Mock Store**

Modify: `internal/pkg/testing/mock/store/store.go`

为 IStore mock 添加 AI 相关方法：
- `AiSessions() AiSessionStore`
- `AiMessages() AiMessageStore`
- `AiUserQuotas() AiUserQuotaStore`

```go
// 在 mock store 中添加
type AiSessionStore struct {
    CreateErr       error
    GetErr          error
    GetResult       *model.AiSessionM
    ListByUIDResult []*model.AiSessionM
    // ... 其他字段
}
```

**Step 2: 定义 ChatBiz 接口**

```go
// internal/apiserver/biz/chat/chat.go

// ABOUTME: Chat business logic interface and implementation.
// ABOUTME: Orchestrates AI chat, session management, and quota control.

package chat

import (
	"context"

	"github.com/bingo-project/bingo/pkg/ai"
)

// ChatBiz defines the chat business interface.
type ChatBiz interface {
	// Chat performs a non-streaming chat completion.
	Chat(ctx context.Context, req *ai.ChatRequest) (*ai.ChatResponse, error)

	// ChatStream performs a streaming chat completion.
	ChatStream(ctx context.Context, req *ai.ChatRequest) (*ai.ChatStream, error)

	// Sessions returns the session management interface.
	Sessions() SessionBiz
}
```

**Step 3: 写 Biz 层测试（使用 mock/store）**

Create: `internal/apiserver/biz/chat/chat_test.go`

使用 `internal/pkg/testing/mock/store` 的 mock 实现：

```go
import mockstore "github.com/bingo-project/bingo/internal/pkg/testing/mock/store"

func TestChatBiz_Chat_Success(t *testing.T) {
    store := mockstore.NewStore()
    // 配置 mock 返回值
    store.AiSessionStore().GetResult = &model.AiSessionM{...}

    biz := chat.New(store, registry)
    // ...
}
```

测试用例：
- TestChatBiz_Chat_Success
- TestChatBiz_Chat_QuotaExceeded
- TestChatBiz_Chat_SessionNotFound
- TestChatBiz_ChatStream_Success

Run: `go test ./internal/apiserver/biz/chat/... -v`
Expected: 测试失败（实现未完成）

**Step 4: 创建 session.go**

实现 SessionBiz 接口，包含会话 CRUD。

**Step 5: 创建 history.go**

实现消息历史拼装和滑动窗口裁剪。

**Step 6: 创建 quota.go**

实现 Token 配额检查和更新。

**Step 7: 实现 chat.go**

实现 ChatBiz 接口。

**Step 8: 修改 biz.go 注入**

在 IBiz 接口添加 Chat() 方法。

**Step 9: 运行测试验证**

Run: `go test ./internal/apiserver/biz/chat/... -v`
Expected: 所有测试通过

**Step 10: Commit**

```bash
make lint
git add internal/pkg/testing/mock/store/store.go internal/apiserver/biz/chat/*.go internal/apiserver/biz/biz.go
git commit -m "feat(ai): add chat business logic"
```

---

## Phase 5: Handler + Router

### Task 5.1: 创建 API DTO

**Files:**
- Create: `pkg/api/apiserver/v1/chat.go`

**Step 1: 创建 chat.go**

定义 ChatCompletionRequest, ChatCompletionResponse, ListSessionsRequest 等 DTO。

**Step 2: Commit**

```bash
make lint
git add pkg/api/apiserver/v1/chat.go
git commit -m "feat(ai): add chat API DTOs"
```

---

### Task 5.2: 创建 Chat Handler

**Files:**
- Modify: `internal/pkg/testing/mock/biz/biz.go` (扩展 Chat Biz mock)
- Create: `internal/apiserver/handler/http/chat/chat.go`
- Create: `internal/apiserver/handler/http/chat/session.go`
- Create: `internal/apiserver/handler/http/chat/stream.go`
- Create: `internal/apiserver/handler/http/chat/chat_test.go`

**Step 1: 扩展 Mock Biz**

Modify: `internal/pkg/testing/mock/biz/biz.go`

为 IBiz mock 添加 Chat 相关方法：
- `Chat() ChatBiz`

```go
// 在 mock biz 中添加
type ChatBiz struct {
    ChatResult       *ai.ChatResponse
    ChatErr          error
    ChatStreamResult *ai.ChatStream
    ChatStreamErr    error
    // ... 其他字段
}
```

**Step 2: 写 Handler 层测试（使用 mock/biz）**

Create: `internal/apiserver/handler/http/chat/chat_test.go`

使用 `internal/pkg/testing/mock/biz` 的 mock 实现：

```go
import mockbiz "github.com/bingo-project/bingo/internal/pkg/testing/mock/biz"

func TestChatHandler_ChatCompletions_Success(t *testing.T) {
    biz := mockbiz.NewBiz()
    biz.ChatBiz().ChatResult = &ai.ChatResponse{...}

    handler := chat.New(biz)
    // ...
}
```

测试用例：
- TestChatHandler_ChatCompletions_Success
- TestChatHandler_ChatCompletions_InvalidRequest
- TestChatHandler_ListModels
- TestChatHandler_ListSessions

Run: `go test ./internal/apiserver/handler/http/chat/... -v`
Expected: 测试失败（实现未完成）

**Step 3: 创建 chat.go**

实现 ChatCompletions, ListModels Handler，包含 Swagger 注释。

**Step 4: 创建 session.go**

实现会话 CRUD Handler。

**Step 5: 创建 stream.go**

实现 SSE 流式响应。

**Step 6: 运行测试验证**

Run: `go test ./internal/apiserver/handler/http/chat/... -v`
Expected: 所有测试通过

**Step 7: Commit**

```bash
make lint
git add internal/pkg/testing/mock/biz/biz.go internal/apiserver/handler/http/chat/*.go
git commit -m "feat(ai): add chat HTTP handlers"
```

---

### Task 5.3: 创建路由和中间件

**Files:**
- Create: `internal/apiserver/router/chat.go`
- Create: `internal/pkg/middleware/http/ai_limiter.go`
- Modify: `internal/apiserver/router/api.go`

**Step 1: 创建 chat.go 路由**

注册 /v1/chat/completions, /v1/models, /v1/ai/sessions 路由。

**Step 2: 创建 ai_limiter.go**

实现按用户 RPM 限流中间件。

**Step 3: 修改 api.go**

在 api.go 中引入 chat 路由。

**Step 4: Commit**

```bash
make lint
git add internal/apiserver/router/chat.go internal/pkg/middleware/http/ai_limiter.go internal/apiserver/router/api.go
git commit -m "feat(ai): add chat router and rate limiter"
```

---

## Phase 6: 初始化和集成

### Task 6.1: 应用初始化

**Files:**
- Modify: `internal/apiserver/app.go`

**Step 1: 初始化 AI Registry**

在 app.go 中根据配置初始化 Registry，注册 Provider。

**Step 2: Commit**

```bash
make lint
git add internal/apiserver/app.go
git commit -m "feat(ai): initialize AI registry in app"
```

---

### Task 6.2: 创建 Seeder (可选)

**Files:**
- Create: `internal/pkg/database/seeder/ai_seeder.go`

**Step 1: 创建初始数据 Seeder**

初始化 ai_quota_tier (free/pro/enterprise) 和常用 ai_provider/ai_model。

**Step 2: Commit**

```bash
make lint
git add internal/pkg/database/seeder/ai_seeder.go
git commit -m "feat(ai): add AI data seeder"
```

---

## Phase 7: 测试和文档

### Task 7.1: 生成 Swagger 文档

**Step 1: 运行 swag**

Run: `make swag`

**Step 2: 验证文档**

访问 http://localhost:8080/swagger/index.html 验证 AI 相关端点。

**Step 3: Commit**

```bash
make lint
git add docs/swagger/*
git commit -m "docs(ai): update swagger documentation"
```

---

### Task 7.2: 端到端测试

**Step 1: 启动服务**

Run: `make build && ./_output/platforms/darwin/arm64/bingo-apiserver -c configs/bingo-apiserver.yaml`

**Step 2: 测试对话接口**

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hello"}]}'
```

**Step 3: 测试流式接口**

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hello"}], "stream": true}'
```

---

## Verification Checklist

- [ ] 6 张表创建成功 (`bingo migrate up`)
- [ ] 配置文件加载正确
- [ ] OpenAI Provider 可正常调用
- [ ] 流式响应正常
- [ ] 会话持久化正常
- [ ] 限流中间件生效
- [ ] Swagger 文档生成

---

## Notes

- 每个文件必须有 2 行 ABOUTME 注释
- Store 放 `internal/pkg/store/`，Model 放 `internal/pkg/model/`
- Handler 放 `internal/apiserver/handler/http/chat/`
- 遵循 TDD：先写测试，再实现
- 每个 Task 完成后 commit
