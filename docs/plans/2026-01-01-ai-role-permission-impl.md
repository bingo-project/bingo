# AI Role Permission Isolation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Split AI role management into admin-only CRUD (admserver) and user read-only (apiserver), enforcing permission isolation through service boundary.

**Architecture:** Move all mutation operations to admserver with existing RBAC protection. Simplify apiserver to only serve active roles to authenticated users. Both services share the same Store layer but have independent Biz layers with different business rules.

**Tech Stack:** Go 1.24+, Gin, GORM, existing RBAC system (Casbin)

---

## Phase 1: Create admserver AI Role Management

### Task 1.1: Create admserver Biz Layer

**Files:**
- Create: `internal/admserver/biz/ai/role.go`

**Step 1: Write the failing test**

Create `internal/admserver/biz/ai/role_test.go`:

```go
package ai

import (
	"context"
	"testing"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAiRoleBiz_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ds := store.NewInMemoryStore()
		biz := NewAiRole(ds)

		req := &v1.CreateAiRoleRequest{
			RoleID:       "test-role",
			Name:         "Test Role",
			Description:  "Test description",
			SystemPrompt: "You are a test role",
			Model:        "gpt-4",
		}

		resp, err := biz.Create(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, "test-role", resp.RoleID)
		assert.Equal(t, "Test Role", resp.Name)
		assert.Equal(t, model.AiRoleStatusActive, resp.Status)
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/admserver/biz/ai/role_test.go -v`
Expected: FAIL with "undefined: NewAiRole"

**Step 3: Write minimal implementation**

Create `internal/admserver/biz/ai/role.go`:

```go
// ABOUTME: AI role business logic for admin management.
// ABOUTME: Provides full CRUD operations for AI role presets with no status restrictions.
package ai

import (
	"context"
	"errors"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

// AiRoleBiz defines AI role management interface for admin.
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

var _ AiRoleBiz = (*aiRoleBiz)(nil)

func NewAiRole(ds store.IStore) AiRoleBiz {
	return &aiRoleBiz{ds: ds}
}

// toRoleInfo converts model.AiRoleM to v1.AiRoleInfo.
func toRoleInfo(m *model.AiRoleM) *v1.AiRoleInfo {
	return &v1.AiRoleInfo{
		RoleID:       m.RoleID,
		Name:         m.Name,
		Description:  m.Description,
		Icon:         m.Icon,
		Category:     string(m.Category),
		SystemPrompt: m.SystemPrompt,
		Model:        m.Model,
		Temperature:  m.Temperature,
		MaxTokens:    m.MaxTokens,
		Sort:         m.Sort,
		Status:       string(m.Status),
	}
}

func (b *aiRoleBiz) Create(ctx context.Context, req *v1.CreateAiRoleRequest) (*v1.AiRoleInfo, error) {
	// Check if role already exists
	existing, err := b.ds.AiRole().GetByRoleID(ctx, req.RoleID)
	if err == nil && existing != nil {
		return nil, errno.ErrResourceAlreadyExists.WithMessage("role_id already exists: %s", req.RoleID)
	}

	// Set default category if not provided
	category := model.AiRoleCategoryGeneral
	if req.Category != "" {
		category = model.AiRoleCategory(req.Category)
	}

	role := &model.AiRoleM{
		RoleID:       req.RoleID,
		Name:         req.Name,
		Description:  req.Description,
		Icon:         req.Icon,
		Category:     category,
		SystemPrompt: req.SystemPrompt,
		Model:        req.Model,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		Sort:         req.Sort,
		Status:       model.AiRoleStatusActive,
	}

	if err := b.ds.AiRole().Create(ctx, role); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("create ai role: %v", err)
	}

	log.C(ctx).Infow("ai role created", "role_id", role.RoleID, "name", role.Name)

	return toRoleInfo(role), nil
}

func (b *aiRoleBiz) Get(ctx context.Context, roleID string) (*v1.AiRoleInfo, error) {
	role, err := b.ds.AiRole().GetByRoleID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIRoleNotFound
		}
		return nil, errno.ErrDBRead.WithMessage("get ai role: %v", err)
	}

	return toRoleInfo(role), nil
}

func (b *aiRoleBiz) List(ctx context.Context, req *v1.ListAiRoleRequest) (*v1.ListAiRoleResponse, error) {
	var category model.AiRoleCategory
	if req.Category != "" {
		category = model.AiRoleCategory(req.Category)
	}

	// Admin can list all statuses, no default filtering
	var status model.AiRoleStatus
	if req.Status != "" {
		status = model.AiRoleStatus(req.Status)
	}

	roles, err := b.ds.AiRole().ListByCategory(ctx, category, status)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("list ai roles: %v", err)
	}

	data := make([]v1.AiRoleInfo, len(roles))
	for i, r := range roles {
		data[i] = *toRoleInfo(r)
	}

	return &v1.ListAiRoleResponse{
		Total: int64(len(roles)),
		Data:  data,
	}, nil
}

func (b *aiRoleBiz) Update(ctx context.Context, roleID string, req *v1.UpdateAiRoleRequest) (*v1.AiRoleInfo, error) {
	role, err := b.ds.AiRole().GetByRoleID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIRoleNotFound
		}
		return nil, errno.ErrDBRead.WithMessage("get ai role: %v", err)
	}

	// Update fields
	_ = copier.CopyWithOption(role, req, copier.Option{IgnoreEmpty: true})
	if req.Category != "" {
		role.Category = model.AiRoleCategory(req.Category)
	}
	if req.Status != "" {
		role.Status = model.AiRoleStatus(req.Status)
	}

	if err := b.ds.AiRole().Update(ctx, role); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("update ai role: %v", err)
	}

	log.C(ctx).Infow("ai role updated", "role_id", role.RoleID)

	return toRoleInfo(role), nil
}

func (b *aiRoleBiz) Delete(ctx context.Context, roleID string) error {
	role, err := b.ds.AiRole().GetByRoleID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrAIRoleNotFound
		}
		return errno.ErrDBRead.WithMessage("get ai role: %v", err)
	}

	role.Status = model.AiRoleStatusDisabled
	if err := b.ds.AiRole().Update(ctx, role, "status"); err != nil {
		return errno.ErrDBWrite.WithMessage("delete ai role: %v", err)
	}

	log.C(ctx).Infow("ai role deleted", "role_id", role.RoleID)

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/admserver/biz/ai/... -v`
Expected: PASS

**Step 5: Register Biz in IBiz**

Modify: `internal/admserver/biz/biz.go`

Add method to interface:
```go
func (b *biz) AiRoles() ai.AiRoleBiz {
	return ai.NewAiRole(b.ds)
}
```

**Step 6: Commit**

```bash
git add internal/admserver/biz/ai/
git commit -m "feat(admserver): add AI role biz layer for admin management"
```

---

### Task 1.2: Create admserver Handler Layer

**Files:**
- Create: `internal/admserver/handler/http/ai/role.go`

**Step 1: Write the handler**

Create `internal/admserver/handler/http/ai/role.go`:

```go
// ABOUTME: HTTP handlers for AI role management in admin panel.
// ABOUTME: Provides CRUD endpoints for AI role resources.
package ai

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/admserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type RoleHandler struct {
	b biz.IBiz
}

func NewRoleHandler(ds store.IStore) *RoleHandler {
	return &RoleHandler{b: biz.NewBiz(ds)}
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
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))
		return
	}

	role, err := h.b.AiRoles().Create(c, &req)
	core.Response(c, role, err)
}

// Get
// @Summary    Get AI role by role_id
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      id  path      string  true  "Role ID"
// @Success    200  {object}  v1.AiRoleInfo
// @Failure    400  {object}  core.ErrResponse
// @Failure    404  {object}  core.ErrResponse
// @Router     /v1/ai/roles/{id} [GET].
func (h *RoleHandler) Get(c *gin.Context) {
	roleID := c.Param("id")
	role, err := h.b.AiRoles().Get(c, roleID)
	core.Response(c, role, err)
}

// List
// @Summary    List AI roles
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      category  query     string  false  "Filter by category"
// @Param      status    query     string  false  "Filter by status"
// @Success    200       {object}  v1.ListAiRoleResponse
// @Failure    400       {object}  core.ErrResponse
// @Router     /v1/ai/roles [GET].
func (h *RoleHandler) List(c *gin.Context) {
	var req v1.ListAiRoleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))
		return
	}

	roles, err := h.b.AiRoles().List(c, &req)
	core.Response(c, roles, err)
}

// Update
// @Summary    Update AI role
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      id        path      string                   true  "Role ID"
// @Param      request  body      v1.UpdateAiRoleRequest   true  "Update request"
// @Success    200      {object}  v1.AiRoleInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    404      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/roles/{id} [PUT].
func (h *RoleHandler) Update(c *gin.Context) {
	var req v1.UpdateAiRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))
		return
	}

	roleID := c.Param("id")
	role, err := h.b.AiRoles().Update(c, roleID, &req)
	core.Response(c, role, err)
}

// Delete
// @Summary    Delete AI role
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      id  path  string  true  "Role ID"
// @Success    200  {object}  nil
// @Failure    404  {object}  core.ErrResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/ai/roles/{id} [DELETE].
func (h *RoleHandler) Delete(c *gin.Context) {
	roleID := c.Param("id")
	err := h.b.AiRoles().Delete(c, roleID)
	core.Response(c, nil, err)
}
```

**Step 2: Register routes**

Modify: `internal/admserver/router/api.go`

Add after line 88 (after role handler section):
```go
// AI Role
aiRoleHandler := ai.NewRoleHandler(store.S)
v1.GET("ai/roles", aiRoleHandler.List)
v1.POST("ai/roles", aiRoleHandler.Create)
v1.GET("ai/roles/:id", aiRoleHandler.Get)
v1.PUT("ai/roles/:id", aiRoleHandler.Update)
v1.DELETE("ai/roles/:id", aiRoleHandler.Delete)
```

Add import:
```go
"github.com/bingo-project/bingo/internal/admserver/handler/http/ai"
```

**Step 3: Commit**

```bash
git add internal/admserver/handler/http/ai/ internal/admserver/router/api.go
git commit -m "feat(admserver): add AI role HTTP handlers and routes"
```

---

## Phase 2: Simplify apiserver to Read-Only

### Task 2.1: Modify apiserver Biz Layer

**Files:**
- Modify: `internal/apiserver/biz/chat/role.go`

**Step 1: Remove mutation methods from interface**

Modify: `internal/apiserver/biz/chat/role.go:20-27`

Replace interface definition:
```go
// AiRoleBiz defines AI role query interface for users.
type AiRoleBiz interface {
	Get(ctx context.Context, roleID string) (*v1.AiRoleInfo, error)
	List(ctx context.Context, req *v1.ListAiRoleRequest) (*v1.ListAiRoleResponse, error)
}
```

**Step 2: Remove mutation methods from implementation**

Delete lines 56-90, 135-181 (Create, Update, Delete methods)

**Step 3: Modify List to force active status**

Modify: `internal/apiserver/biz/chat/role.go:105-133`

Replace List method:
```go
func (b *aiRoleBiz) List(ctx context.Context, req *v1.ListAiRoleRequest) (*v1.ListAiRoleResponse, error) {
	var category model.AiRoleCategory
	if req.Category != "" {
		category = model.AiRoleCategory(req.Category)
	}

	// Force filter to active only, ignore status parameter
	status := model.AiRoleStatusActive

	roles, err := b.ds.AiRole().ListByCategory(ctx, category, status)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("list ai roles: %v", err)
	}

	data := make([]v1.AiRoleInfo, len(roles))
	for i, r := range roles {
		data[i] = *toRoleInfo(r)
	}

	return &v1.ListAiRoleResponse{
		Total: int64(len(roles)),
		Data:  data,
	}, nil
}
```

**Step 4: Modify Get to only return active roles**

Modify: `internal/apiserver/biz/chat/role.go:92-103`

Replace Get method:
```go
func (b *aiRoleBiz) Get(ctx context.Context, roleID string) (*v1.AiRoleInfo, error) {
	role, err := b.ds.AiRole().GetByRoleID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIRoleNotFound
		}
		return nil, errno.ErrDBRead.WithMessage("get ai role: %v", err)
	}

	// Only return active roles
	if role.Status != model.AiRoleStatusActive {
		return nil, errno.ErrAIRoleNotFound
	}

	return toRoleInfo(role), nil
}
```

**Step 5: Commit**

```bash
git add internal/apiserver/biz/chat/role.go
git commit -m "refactor(apiserver): simplify AI role biz to read-only, force active filter"
```

---

### Task 2.2: Remove Mutation Handlers from apiserver

**Files:**
- Modify: `internal/apiserver/handler/http/chat/role.go`
- Modify: `internal/apiserver/router/ai.go`

**Step 1: Remove mutation methods**

Delete lines 27-48, 89-113, 116-131 (Create, Update, Delete methods and Swagger)

**Step 2: Remove status parameter from List Swagger**

Modify: `internal/apiserver/handler/http/chat/role.go:67-77`

Replace Swagger annotation:
```go
// List
// @Summary    List AI roles
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      category  query     string  false  "Filter by category"
// @Success    200       {object}  v1.ListAiRoleResponse
// @Failure    400       {object}  core.ErrResponse
// @Router     /v1/ai/roles [GET].
func (h *RoleHandler) List(c *gin.Context) {
	var req v1.ListAiRoleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))
		return
	}

	roles, err := h.b.AiRoles().List(c, &req)
	core.Response(c, roles, err)
}
```

**Step 3: Remove mutation routes**

Modify: `internal/apiserver/router/ai.go:35-43`

Replace route registration:
```go
// Role presets (read-only for users)
roles := g.Group("/ai/roles")
{
	roles.GET("", roleHandler.List)
	roles.GET("/:role_id", roleHandler.Get)
}
```

**Step 4: Commit**

```bash
git add internal/apiserver/handler/http/chat/role.go internal/apiserver/router/ai.go
git commit -m "refactor(apiserver): remove AI role mutation handlers and routes"
```

---

## Phase 3: Add API Seeder Records

### Task 3.1: Add AI Role APIs to seeder

**Files:**
- Modify: `internal/pkg/database/seeder/api_seeder.go`

**Step 1: Add AI role APIs to coreAPIs**

Modify: `internal/pkg/database/seeder/api_seeder.go:14-92`

Add after line 91 (after File management):
```go
	// AI Role management
	{Method: "GET", Path: "/v1/ai/roles", Group: "AI", Description: "List AI roles"},
	{Method: "POST", Path: "/v1/ai/roles", Group: "AI", Description: "Create AI role"},
	{Method: "GET", Path: "/v1/ai/roles/:id", Group: "AI", Description: "Get AI role"},
	{Method: "PUT", Path: "/v1/ai/roles/:id", Group: "AI", Description: "Update AI role"},
	{Method: "DELETE", Path: "/v1/ai/roles/:id", Group: "AI", Description: "Delete AI role"},
```

**Step 2: Run seeder**

Run: `bingo db seed --seeder=ApiSeeder`
Expected: Success message, 5 new API records created

**Step 3: Verify in database**

Run:
```bash
mysql -u root -p bingo_dev -e "SELECT method, path, \`group\`, description FROM api WHERE \`group\` = 'AI';"
```
Expected: 5 rows with AI role endpoints

**Step 4: Commit**

```bash
git add internal/pkg/database/seeder/api_seeder.go
git commit -m "feat(seeder): add AI role management APIs"
```

---

## Phase 4: Update Swagger Documentation

### Task 4.1: Regenerate Swagger docs

**Step 1: Run swag**

Run: `make swag`
Expected: Success message, docs regenerated

**Step 2: Commit**

```bash
git add api/swagger/
git commit -m "docs(swagger): regenerate API documentation for AI role changes"
```

---

## Phase 5: Build and Test

### Task 5.1: Build services

**Step 1: Build all services**

Run: `make build`
Expected: All binaries compiled successfully

**Step 2: Commit**

```bash
git commit --allow-empty -m "build: successful compilation of AI role permission isolation"
```

---

### Task 5.2: Integration testing

**Step 1: Test apiserver read-only**

```bash
# Start apiserver
./_output/platforms/darwin/arm64/bingo-apiserver &

# Test as normal user
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"account":"user","password":"pass"}' | jq -r '.data.accessToken')

# List roles (should work)
curl -s http://localhost:8080/v1/ai/roles \
  -H "Authorization: Bearer $TOKEN" | jq

# Try to create role (should 404)
curl -s -X POST http://localhost:8080/v1/ai/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role_id":"test","name":"Test"}' | jq
```

Expected: List returns active roles, Create returns 404

**Step 2: Test admserver CRUD**

```bash
# Start admserver
./_output/platforms/darwin/arm64/bingo-admserver &

# Test as admin
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"account":"admin","password":"admin123"}' | jq -r '.data.accessToken')

# Create role (should work)
curl -s -X POST http://localhost:8081/v1/ai/roles \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role_id":"test-role",
    "name":"Test Role",
    "description":"Test",
    "system_prompt":"You are helpful",
    "model":"gpt-4"
  }' | jq

# List all roles (should work)
curl -s http://localhost:8081/v1/ai/roles \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

Expected: All CRUD operations succeed

**Step 3: Test status filtering**

```bash
# List disabled roles (admin only)
curl -s "http://localhost:8081/v1/ai/roles?status=disabled" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Try on apiserver (should be ignored, only active returned)
curl -s "http://localhost:8080/v1/ai/roles?status=disabled" \
  -H "Authorization: Bearer $TOKEN" | jq
```

Expected: admserver returns disabled, apiserver ignores parameter

**Step 4: Commit**

```bash
git commit --allow-empty -m "test: integration tests pass for AI role permission isolation"
```

---

## Verification Checklist

- [ ] admserver has full CRUD handlers for AI roles
- [ ] apiserver only has List and Get handlers
- [ ] apiserver Biz layer forces `status=active` filter
- [ ] apiserver Biz layer returns 404 for non-active roles
- [ ] API seeder includes AI role endpoints
- [ ] Swagger documentation updated
- [ ] All services build successfully
- [ ] Integration tests pass
- [ ] Normal users cannot create/update/delete roles
- [ ] Admins can manage all roles through admserver

---

## References

- Design document: [docs/plans/2026-01-01-ai-role-permission-design.md](../docs/plans/2026-01-01-ai-role-permission-design.md)
- CONVENTIONS: [docs/guides/CONVENTIONS.md](../docs/guides/CONVENTIONS.md)
- Existing AI role biz: [internal/apiserver/biz/chat/role.go](../internal/apiserver/biz/chat/role.go)
- Example admin handler: [internal/admserver/handler/http/system/role.go](../internal/admserver/handler/http/system/role.go)
- API seeder pattern: [internal/pkg/database/seeder/api_seeder.go](../internal/pkg/database/seeder/api_seeder.go)
