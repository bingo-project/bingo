# AI Role Presets Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement AI Role Presets to allow users to switch between different AI personas (e.g., Teacher, Doctor) with customized system prompts.

**Architecture:** Layered architecture (Handler -> Biz -> Store). Store manages `ai_role` table. Biz logic handles role retrieval and validation. Chat Biz integrates role injection into the message stream.

**Tech Stack:** Go, GORM, Gin, MySQL.

### Task 1: Model & Store Layer

**Files:**
- Create: `internal/pkg/model/ai_role.go`
- Create: `internal/pkg/store/ai_role.go`
- Test: `internal/pkg/store/ai_role_test.go` (if feasible to mock DB, or skip if project relies on integration tests for store)
- Modify: `internal/pkg/store/store.go` (Add AiRole interface to IStore)

**Step 1: Define Model**

Create `internal/pkg/model/ai_role.go` with `AiRoleM` struct matching the schema design. Include `TableName()` method.

**Step 2: Define Store Interface & Implementation**

Create `internal/pkg/store/ai_role.go`:
- Define `AiRoleStore` interface (Create, Update, Delete, Get, List).
- Implement `aiRoleStore` struct using `genericstore`.
- Implement `GetByRoleID(ctx, roleID)` method.

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
- Create: `internal/pkg/database/migration/20251231120000_create_ai_role_table.go` (approx name)
- Modify: `internal/pkg/errno/ai.go`
- Modify: `pkg/api/apiserver/v1/chat.go` (Add RoleID to ChatRequest)

**Step 1: Create Migration**

Create the migration file using `gorm` auto-migrate or raw SQL as per project pattern. Use `AiRoleM` model if auto-migrate is used, or raw SQL from design doc.

**Step 2: Add Constants & Errors**

- In `internal/pkg/errno/ai.go`, add `ErrAIRoleNotFound` and `ErrAIRoleDisabled`.
- In `internal/pkg/model/ai_role.go` (or `known` pkg), add constants for Status (`active`, `disabled`) and Category.

**Step 3: Update Chat API DTO**

In `pkg/api/apiserver/v1/chat.go`, add `RoleID string` field to `ChatRequest`.

**Step 4: Commit**

```bash
git add internal/pkg/database/migration/ internal/pkg/errno/ai.go pkg/api/apiserver/v1/chat.go internal/pkg/model/ai_role.go
git commit -m "feat(ai): add database migration and error codes for AI roles"
```

### Task 3: Biz Layer (Role Management)

**Files:**
- Create: `internal/apiserver/biz/role/role.go`
- Modify: `internal/apiserver/biz/biz.go`
- Test: `internal/apiserver/biz/role/role_test.go`

**Step 1: Define Biz Interface**

In `internal/apiserver/biz/role/role.go`, define `IBiz` subset or `RoleBiz` interface:
- `Create(ctx, req)`
- `Get(ctx, id)`
- `List(ctx, req)`
- `Update(ctx, id, req)`
- `Delete(ctx, id)`

**Step 2: Implement Biz Logic**

Implement `roleBiz` struct. Ensure it uses `errno` for errors and logs operations.

**Step 3: Register in IBiz**

Modify `internal/apiserver/biz/biz.go` to include `AiRoles() role.RoleBiz`.

**Step 4: Commit**

```bash
git add internal/apiserver/biz/role/ internal/apiserver/biz/biz.go
git commit -m "feat(ai): add AI role business logic"
```

### Task 4: Integration with Chat Biz

**Files:**
- Modify: `internal/apiserver/biz/chat/chat.go`

**Step 1: Implement Helper Method**

Add `buildMessagesWithRole(ctx, req)` to `chatBiz`.
- Fetch role by ID.
- Validate status.
- Prepend system prompt.
- Override model/temp/max_tokens if set in role.

**Step 2: Integrate in Chat**

Call `buildMessagesWithRole` at the start of `Chat` and `ChatStream` methods.

**Step 3: Commit**

```bash
git add internal/apiserver/biz/chat/chat.go
git commit -m "feat(ai): integrate role presets into chat flow"
```

### Task 5: API & Handler Layer

**Files:**
- Create: `pkg/api/apiserver/v1/role.go`
- Create: `internal/apiserver/handler/http/role/role.go`
- Modify: `internal/apiserver/router/router.go`

**Step 1: Define DTOs**

In `pkg/api/apiserver/v1/role.go`, define request/response structs for Role CRUD (Create, Update, List, Get).

**Step 2: Implement HTTP Handler**

Create `internal/apiserver/handler/http/role/role.go`.
- Implement `Create`, `Get`, `List`, `Update`, `Delete` methods.
- **IMPORTANT**: Add Swagger comments for each method.

**Step 3: Register Routes**

In `internal/apiserver/router/router.go`:
- Register `/v1/ai/roles` group.
- Add routes mapping to handler methods.
- Apply auth middleware if needed (POST/PUT/DELETE usually admin only).

**Step 4: Commit**

```bash
git add pkg/api/apiserver/v1/role.go internal/apiserver/handler/http/role/role.go internal/apiserver/router/router.go
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
