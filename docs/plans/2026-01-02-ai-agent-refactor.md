# AI Agent Refactoring & Chat API Update Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Rename "AI Role" to "AI Agent" across the codebase to align with industry terminology and enable future agentic features. Make `model` parameter optional in Chat API to support both stateless and session-based interactions.

**Architecture:** 
- **Refactoring:** Rename DTOs, Models, Stores, and Handlers from `Role` to `Agent`. Update API routes.
- **Database:** Rename table `ai_roles` to `ai_agents`. Rename column `role_id` to `agent_id`.
- **Logic:** Update Chat logic to fallback to Agent's model if request model is empty.

**Tech Stack:** Go, Gin, GORM.

### Task 1: Rename API DTOs (pkg/api)

**Files:**
- Rename: `pkg/api/apiserver/v1/ai_role.go` -> `pkg/api/apiserver/v1/ai_agent.go`
- Modify: `pkg/api/apiserver/v1/ai_chat.go`
- Modify: `pkg/api/apiserver/v1/ai_agent.go`

**Step 1: Rename file and structs in ai_agent.go**
- Rename `ai_role.go` to `ai_agent.go`.
- Rename `CreateAiRoleRequest` -> `CreateAiAgentRequest`.
- Rename `UpdateAiRoleRequest` -> `UpdateAiAgentRequest`.
- Rename `AiRoleInfo` -> `AiAgentInfo`.
- Rename `ListAiRoleRequest/Response` -> `ListAiAgentRequest/Response`.
- Update JSON tags: `roleId` -> `agentId`.

**Step 2: Update References in ai_chat.go**
- Update `CreateSessionRequest`: `RoleID` -> `AgentID` (`json:"agentId"`).
- Update `SessionInfo`: `RoleID` -> `AgentID`, `RoleName` -> `AgentName`.
- Make `Model` in `ChatCompletionRequest` optional (`omitempty`, remove binding required).

**Step 3: Verification**
- Run `go build ./pkg/api/...` to check for compilation errors (will fail until other packages updated, so maybe skip build here or just check file syntax).

### Task 2: Update Data Model (internal/pkg/model)

**Files:**
- Rename: `internal/pkg/model/ai_role.go` -> `internal/pkg/model/ai_agent.go`

**Step 1: Rename Struct and TableName**
- Rename struct `AiRole` -> `AiAgent`.
- Update `TableName()` to return `ai_agents`.
- Update gorm tags or fields if necessary (`RoleID` -> `AgentID`).

### Task 3: Update Store Layer (internal/pkg/store)

**Files:**
- Rename: `internal/pkg/store/ai_role.go` -> `internal/pkg/store/ai_agent.go`
- Modify: `internal/pkg/store/ai_agent.go`
- Modify: `internal/pkg/store/store.go` (Interface definition)

**Step 1: Rename Interface**
- Rename `AiRoleStore` -> `AiAgentStore`.
- Update methods: `CreateAiRole` -> `CreateAiAgent`, etc.

**Step 2: Update Implementation**
- Update `ai_agent.go` implementation to use `AiAgent` model and new interface methods.

**Step 3: Update Store Registry**
- In `store.go`, rename `AiRoles()` -> `AiAgents()`.

### Task 4: Database Migration

**Files:**
- Create: `internal/pkg/database/migration/2026_01_02_rename_ai_roles_to_agents.go`

**Step 1: Create Migration**
- Rename table `ai_roles` to `ai_agents`.
- Rename column `role_id` to `agent_id` in `ai_agents` table.
- Rename column `role_id` to `agent_id` in `chat_sessions` table (if exists and linked).

### Task 5: Update Admin Handler & Biz (internal/admserver)

**Files:**
- Rename: `internal/admserver/handler/http/ai/role.go` -> `internal/admserver/handler/http/ai/agent.go`
- Rename: `internal/admserver/biz/ai/role.go` -> `internal/admserver/biz/ai/agent.go`
- Modify: `internal/admserver/biz/biz.go`
- Modify: `internal/admserver/router/api.go`

**Step 1: Update Biz Layer**
- Rename `AiRoleBiz` -> `AiAgentBiz`.
- Update `biz.go` interface.

**Step 2: Update Handler**
- Rename `AiRoleHandler` -> `AiAgentHandler`.
- Update logic to use `AiAgentBiz`.

**Step 3: Update Router**
- Change route group `/ai/roles` to `/ai/agents`.
- Map new handler methods.

### Task 6: Update Chat Handler & Logic (internal/apiserver)

**Files:**
- Modify: `internal/apiserver/handler/http/chat/chat.go`
- Modify: `internal/apiserver/biz/chat/chat.go`
- Modify: `internal/apiserver/biz/chat/session.go`

**Step 1: Implement Model Fallback Logic in Handler**
- In `ChatCompletions`, check if `req.Model` is empty.
- If empty, resolve `SessionID`.
- In `Biz` layer, if Session exists, fetch `Agent` to get `DefaultModel`.
- Use Agent's model if request model is empty.

**Step 2: Update Session Creation Logic**
- Update `CreateSession` to accept `AgentID` instead of `RoleID`.

### Task 7: Update Swagger Docs

**Files:**
- Run command: `make swag`

**Step 1: Regenerate Docs**
- Run `make swag` to reflect API changes.

### Task 8: Manual Verification

**Step 1: Test Stateless Chat**
- `curl /v1/chat/completions` with explicit `model`.

**Step 2: Test Agent Chat**
- Create Agent (`/v1/ai/agents`).
- Create Session (`/v1/chat/sessions`) with `agentId`.
- `curl /v1/chat/completions` without `model`, with `sessionId`.
