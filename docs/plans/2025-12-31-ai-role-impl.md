# AI Role Presets Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement AI Role Presets to allow users to switch between different AI personas (e.g., Teacher, Doctor) with customized system prompts.

**Architecture:** Layered architecture (Handler -> Biz -> Store). Store manages `ai_role` table. Biz logic handles role retrieval and validation. Chat Biz integrates role injection into the message stream.

**Tech Stack:** Go, GORM, Gin, MySQL.

---

## Pre-Implementation Checklist

Before starting implementation, ensure:

- [ ] Read CONVENTIONS.md and understand three-layer architecture
- [ ] Each new .go file MUST start with 2-line ABOUTME comment
- [ ] Model constants MUST use typed constants (not plain string)
- [ ] All layers require test files
- [ ] Key operations require structured logging
- [ ] Handler methods require Swagger comments

---

## Implementation Checklist

### Task 1: Model & Store Layer

**Files:**
- Create: `internal/pkg/model/ai_role.go`
- Create: `internal/pkg/store/ai_role.go`
- Test: `internal/pkg/store/ai_role_test.go` (if feasible to mock DB, or skip if project relies on integration tests for store)
- Modify: `internal/pkg/store/store.go` (Add AiRole interface to IStore)

**Step 1: Define Model**

Create `internal/pkg/model/ai_role.go` with `AiRoleM` struct matching the schema design. Include `TableName()` method.

**IMPORTANT:**
- File MUST start with 2-line ABOUTME comment
- Status and Category MUST use typed constants (not plain string)

```go
// ABOUTME: AI role preset model definition.
// ABOUTME: Represents AI persona configurations with custom system prompts.
package model

// AiRoleStatus represents the status of an AI role.
type AiRoleStatus string

const (
    AiRoleStatusActive   AiRoleStatus = "active"
    AiRoleStatusDisabled AiRoleStatus = "disabled"
)

// AiRoleCategory represents the category of an AI role.
type AiRoleCategory string

const (
    AiRoleCategoryGeneral   AiRoleCategory = "general"
    AiRoleCategoryEducation AiRoleCategory = "education"
    AiRoleCategoryMedical   AiRoleCategory = "medical"
    AiRoleCategoryWorkplace AiRoleCategory = "workplace"
    AiRoleCategoryCreative  AiRoleCategory = "creative"
)

// AiRoleM represents an AI role preset.
type AiRoleM struct {
    ID           uint64            `gorm:"primaryKey"`
    RoleID       string            `gorm:"uniqueIndex;size:32;not null"`
    Name         string            `gorm:"size:64;not null"`
    Description  string            `gorm:"size:255"`
    Icon         string            `gorm:"size:255"`
    Category     AiRoleCategory    `gorm:"size:32;default:'general'"`
    SystemPrompt string            `gorm:"type:text;not null"`
    Model        string            `gorm:"size:64"`
    Temperature  float64           `gorm:"type:decimal(3,2);default:0.70"`
    MaxTokens    int               `gorm:"default:2000"`
    Sort         int               `gorm:"default:0"`
    Status       AiRoleStatus      `gorm:"size:16;default:'active'"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

func (AiRoleM) TableName() string {
    return "ai_role"
}
```

**Step 2: Define Store Interface & Implementation**

Create `internal/pkg/store/ai_role.go`:
- Define `AiRoleStore` interface (Create, Update, Delete, Get, List).
- Implement `aiRoleStore` struct using `genericstore`.
- Implement `GetByRoleID(ctx, roleID)` method.
- Add `AiRoleExpansion` interface for custom methods like `GetByRoleID`.

```go
// ABOUTME: AI role data access layer.
// ABOUTME: Provides CRUD operations for AI role preset records.
package store

type AiRoleStore interface {
    Create(ctx context.Context, role *model.AiRoleM) error
    Update(ctx context.Context, role *model.AiRoleM) error
    Delete(ctx context.Context, id uint64) error
    Get(ctx context.Context, opts *where.Options) (*model.AiRoleM, error)
    List(ctx context.Context, opts *where.Options) ([]*model.AiRoleM, int64, error)

    AiRoleExpansion
}

type AiRoleExpansion interface {
    GetByRoleID(ctx context.Context, roleID string) (*model.AiRoleM, error)
    ListByCategory(ctx context.Context, category model.AiRoleCategory, status model.AiRoleStatus) ([]*model.AiRoleM, error)
}
```

**Step 2.5: Create Store Test File**

Create `internal/pkg/store/ai_role_test.go` with basic tests for `GetByRoleID` and `ListByCategory`.

**Step 3: Register in IStore**

Modify `internal/pkg/store/store.go` to include `AiRoles() AiRoleStore`.
Modify `internal/pkg/store/datastore.go` to initialize it.

**Step 4: Commit**

```bash
git add internal/pkg/model/ai_role.go internal/pkg/store/ai_role.go internal/pkg/store/store.go internal/pkg/store/datastore.go
git commit -m "feat(ai): add AiRole model and store implementation"
```

### Task 2: Migration & Constants

**Files:**
- Create: `internal/pkg/database/migration/20260101_create_ai_role_table.go`
- Modify: `internal/pkg/errno/ai.go`
- Modify: `pkg/api/apiserver/v1/chat.go` (Add RoleID to ChatCompletionRequest)

**Step 1: Create Migration**

Create the migration file using `gorm` auto-migrate or raw SQL as per project pattern. Use `AiRoleM` model if auto-migrate is used, or raw SQL from design doc.

**Step 2: Add Error Codes**

In `internal/pkg/errno/ai.go`, add error codes (使用现有的 errorsx.ErrorX 格式):

```go
	// ErrAIRoleNotFound 角色不存在
	ErrAIRoleNotFound = &errorsx.ErrorX{
		Code:    http.StatusNotFound,
		Reason:  "NotFound.AIRoleNotFound",
		Message: "AI role not found.",
	}

	// ErrAIRoleDisabled 角色已禁用
	ErrAIRoleDisabled = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AIRoleDisabled",
		Message: "AI role is disabled.",
	}
```

**Note:** Status and Category constants are already defined in `model/ai_role.go` from Task 1. Do NOT redefine them elsewhere.

**Step 3: Update Chat API DTO**

In `pkg/api/apiserver/v1/chat.go`, add `RoleID string` field to `ChatRequest`.

**Step 4: Commit**

```bash
git add internal/pkg/database/migration/ internal/pkg/errno/ai.go pkg/api/apiserver/v1/chat.go internal/pkg/model/ai_role.go
git commit -m "feat(ai): add database migration and error codes for AI roles"
```

### Task 3: Biz Layer (Role Management)

**Files:**
- Create: `internal/apiserver/biz/chat/role.go` (放在 chat 子包中)
- Modify: `internal/apiserver/biz/biz.go`
- Test: `internal/apiserver/biz/chat/role_test.go`

**Step 1: Define Biz Interface**

In `internal/apiserver/biz/chat/role.go`, define `AiRoleBiz` interface:
- `Create(ctx, req)`
- `Get(ctx, roleID)` (使用 role_id 而非 id)
- `List(ctx, req)`
- `Update(ctx, id, req)`
- `Delete(ctx, id)`

**Step 2: Implement Biz Logic**

Implement `roleBiz` struct. Ensure it:
- Uses `errno` for errors (never return raw Store errors)
- Uses structured logging for key operations

```go
// ABOUTME: AI role business logic.
// ABOUTME: Handles role CRUD operations and validations.
package chat

import (
    "context"

    "<module>/internal/pkg/log"
    "<module>/internal/pkg/store"
    v1 "<module>/pkg/api/apiserver/v1"
)

type AiRoleBiz interface {
    Create(ctx context.Context, req *v1.CreateAiRoleRequest) (*v1.AiRoleInfo, error)
    Get(ctx context.Context, roleID string) (*v1.AiRoleInfo, error)
    List(ctx context.Context, req *v1.ListAiRoleRequest) (*v1.ListAiRoleResponse, error)
    Update(ctx context.Context, roleID string, req *v1.UpdateAiRoleRequest) (*v1.AiRoleInfo, error)
    Delete(ctx context.Context, roleID string) error
}

type aiRoleBiz struct {
    ds store.IStore
}

func NewAiRole(ds store.IStore) AiRoleBiz {
    return &aiRoleBiz{ds: ds}
}

func (b *aiRoleBiz) Create(ctx context.Context, req *v1.CreateAiRoleRequest) (*v1.AiRoleInfo, error) {
    // 1. Business validation
    if exists, _ := b.ds.AiRoles().GetByRoleID(ctx, req.RoleID); exists != nil {
        return nil, errno.ErrAIRoleAlreadyExist
    }

    // 2. Create model
    role := &model.AiRoleM{
        RoleID:       req.RoleID,
        Name:         req.Name,
        Description:  req.Description,
        Category:     model.AiRoleCategory(req.Category),
        SystemPrompt: req.SystemPrompt,
        Status:       model.AiRoleStatusActive,
    }

    // 3. Persist
    if err := b.ds.AiRoles().Create(ctx, role); err != nil {
        return nil, errno.ErrDBWrite.WithMessage("create ai role: %v", err)
    }

    log.C(ctx).Infow("ai role created", "role_id", role.RoleID, "name", role.Name)
    return b.toRoleInfo(role), nil
}
```

**Step 3: Register in IBiz**

Modify `internal/apiserver/biz/biz.go` to include `AiRoles() chat.AiRoleBiz`.

**Step 4: Commit**

```bash
git add internal/apiserver/biz/chat/role.go internal/apiserver/biz/biz.go
git commit -m "feat(ai): add AI role business logic"
```

### Task 4: Integration with Chat Biz

**Files:**
- Modify: `internal/apiserver/biz/chat/chat.go`

**Step 1: Implement Helper Method**

Add `buildMessagesWithRole(ctx, req)` to `chatBiz`:
- Fetch role by ID
- Validate status (must be `model.AiRoleStatusActive`)
- Prepend system prompt
- Override model/temp/max_tokens if set in role
- Log on role load/failure

```go
func (b *chatBiz) buildMessagesWithRole(ctx context.Context, req *ai.ChatRequest) ([]ai.Message, error) {
    if req.RoleID == "" {
        return req.Messages, nil
    }

    role, err := b.ds.AiRoles().GetByRoleID(ctx, req.RoleID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errno.ErrAIRoleNotFound
        }
        return nil, errno.ErrDBRead.WithMessage("get ai role: %v", err)
    }

    // Use typed constant for comparison
    if role.Status != model.AiRoleStatusActive {
        log.C(ctx).Warnw("ai role is disabled", "role_id", req.RoleID)
        return nil, errno.ErrAIRoleDisabled
    }

    // Prepend system prompt
    messages := []ai.Message{
        {Role: ai.RoleSystem, Content: role.SystemPrompt},
    }
    messages = append(messages, req.Messages...)

    // Override model/params if role specifies them
    if role.Model != "" {
        req.Model = role.Model
    }
    if role.Temperature > 0 {
        req.Temperature = &role.Temperature
    }
    if role.MaxTokens > 0 {
        req.MaxTokens = role.MaxTokens
    }

    log.C(ctx).Debugw("ai role loaded", "role_id", req.RoleID, "model", req.Model)
    return messages, nil
}
```

**Step 2: Integrate in Chat**

Call `buildMessagesWithRole` at the start of `Chat` and `ChatStream` methods.

**Step 3: Commit**

```bash
git add internal/apiserver/biz/chat/chat.go
git commit -m "feat(ai): integrate role presets into chat flow"
```

### Task 5: API & Handler Layer

**Files:**
- Create: `pkg/api/apiserver/v1/ai_role.go`
- Create: `internal/apiserver/handler/http/chat/role.go` (放在 chat handler 包中)
- Rename: `internal/apiserver/router/chat.go` → `internal/apiserver/router/ai.go`

**Step 1: Define DTOs**

In `pkg/api/apiserver/v1/ai_role.go`, define request/response structs for AI Role CRUD (CreateAiRoleRequest, UpdateAiRoleRequest, ListAiRoleRequest, AiRoleInfo, ListAiRoleResponse).

**Step 2: Implement HTTP Handler**

Create `internal/apiserver/handler/http/chat/role.go`:
- Implement `Create`, `Get`, `List`, `Update`, `Delete` methods
- **CRITICAL**: Each method MUST have Swagger comments
- Use `core.Response()` for all responses

```go
// ABOUTME: HTTP handlers for AI role preset management.
// ABOUTME: Provides CRUD endpoints for AI role preset resources.
package chat

import (
    "github.com/gin-gonic/gin"

    "<module>/internal/apiserver/biz"
    "<module>/internal/pkg/core"
)

type RoleHandler struct {
    biz biz.IBiz
}

func NewRoleHandler(biz biz.IBiz) *RoleHandler {
    return &RoleHandler{biz: biz}
}

// Create
// @Summary    Create AI role
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      request  body      v1.CreateAiRoleRequest  true  "Param"
// @Success    200      {object}  v1.AiRoleInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/roles [POST].
func (h *RoleHandler) Create(c *gin.Context) {
    var req v1.CreateAiRoleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
        return
    }

    role, err := h.biz.AiRoles().Create(c, &req)
    core.Response(c, role, err)
}

// Get
// @Summary    Get AI role by ID
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      role_id  path      string  true  "Role ID"
// @Success    200      {object}  v1.AiRoleInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    404      {object}  core.ErrResponse
// @Router     /v1/ai/roles/{role_id} [GET].
func (h *RoleHandler) Get(c *gin.Context) {
    // ... implementation
}

// List
// @Summary    List AI roles
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      category  query     string  false  "Filter by category"
// @Param      status    query     string  false  "Filter by status"
// @Param      page      query     int     false  "Page number"
// @Param      page_size query     int     false  "Page size"
// @Success    200       {object}  v1.ListAiRoleResponse
// @Failure    400       {object}  core.ErrResponse
// @Router     /v1/ai/roles [GET].
func (h *RoleHandler) List(c *gin.Context) {
    // ... implementation
}

// Similar for Update and Delete...
```

**Step 3: Rename and Update Router**

Rename `internal/apiserver/router/chat.go` to `internal/apiserver/router/ai.go`:
- Update ABOUTME comments
- Register `/v1/ai/roles` group
- Add routes mapping to handler methods
- Apply auth middleware (POST/PUT/DELETE admin only, GET public)

**Step 4: Commit**

```bash
git add pkg/api/apiserver/v1/ai_role.go internal/apiserver/handler/http/chat/role.go
git mv internal/apiserver/router/chat.go internal/apiserver/router/ai.go
git commit -m "feat(ai): add AI role API handlers and routes"
```

### Task 6: Final Verification

**Files:**
- None

**Step 1: Generate Swagger**

Run `make swag`.

**Step 2: Build**

Run `make build`.

**Step 3: Lint**

Run `make lint`.

**Step 4: Commit generated files**

```bash
git add docs/api/ internal/apiserver/docs/
git commit -m "docs: update swagger documentation for AI roles"
```

---

## Final Checklist (Before Merge)

### CONVENTIONS.md Compliance

- [ ] **Three-layer architecture**: No cross-layer calls (Handler → Biz → Store only)
- [ ] **File headers**: All new .go files have 2-line ABOUTME comment
- [ ] **Typed constants**: `AiRoleStatus` and `AiRoleCategory` are typed, not plain strings
- [ ] **Store naming**: `AiRoleStore` interface, `aiRoleStore` struct, `AiRoleExpansion`
- [ ] **Error handling**: Biz layer wraps all Store errors with `errno`
- [ ] **Logging**: Structured logging (`log.C(ctx).Infow()`) for key operations
- [ ] **Swagger**: All HTTP handler methods have complete Swagger comments
- [ ] **Context**: Handler passes `*gin.Context` (not `c.Request.Context()`) to Biz

### Testing

- [ ] Store layer test: `internal/pkg/store/ai_role_test.go`
- [ ] Biz layer test: `internal/apiserver/biz/role/role_test.go`
- [ ] Handler layer test: `internal/apiserver/handler/http/role/role_test.go`
- [ ] Tests use proper mocking (Store → SQLite, Biz/Handler → mocks)

### Database

- [ ] Migration file created and tested (`bingo migrate up`)
- [ ] Migration supports rollback (`bingo migrate rollback`)
- [ ] Seeder prepared for sample data (optional but recommended)

### API

- [ ] All CRUD endpoints work (Create, Get, List, Update, Delete)
- [ ] Role filtering by category works
- [ ] Chat endpoint respects `role_id` parameter
- [ ] Disabled roles are rejected
- [ ] Non-existent roles return proper error
