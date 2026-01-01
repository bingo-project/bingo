# AI Role 权限隔离设计

**日期**: 2026-01-01
**状态**: 设计阶段

## 概述

将 AI Role 管理从单一服务拆分为后台管理和前端用户查询，实现权限隔离。

**当前问题**：
- AI Role CRUD 接口都在 apiserver，只做了认证未做授权
- 任何登录用户都可以创建/修改/删除 AI Role
- 注释声称 "public GET, admin-only mutations" 但未实现

**设计目标**：
- 管理员通过 admserver 统一管理 AI Role
- 前端用户只能查询 active 状态的角色
- 利用现有 RBAC 系统，无需重复造轮子

## 接口设计

### admserver（后台管理）

```
GET    /v1/ai/roles          # 列表（支持筛选所有状态）
POST   /v1/ai/roles          # 创建
GET    /v1/ai/roles/:id      # 详情
PUT    /v1/ai/roles/:id      # 更新
DELETE /v1/ai/roles/:id      # 删除
```

**权限**：通过 Casbin 策略控制，仅管理员可访问

### apiserver（前端用户）

```
GET /v1/ai/roles             # 查询 active 状态的角色
GET /v1/ai/roles/:role_id    # 查询 active 状态的角色详情
```

**权限**：所有登录用户可访问，业务层强制过滤 `status=active`

## 实现细节

### 代码组织

**admserver**:
```
internal/admserver/
├── handler/http/ai/role.go      # CRUD Handlers
├── biz/ai/role.go               # 完整业务逻辑
└── router/api.go                # 注册路由
```

**apiserver**:
```
internal/apiserver/
├── handler/http/chat/role.go    # 只保留 Get/List
├── biz/chat/role.go             # 只读逻辑，强制 status=active
└── router/ai.go                 # 移除 POST/PUT/DELETE
```

**共享**:
```
internal/pkg/store/ai_role.go   # 数据访问层（不变）
pkg/api/apiserver/v1/ai_role.go # API 定义（不变）
```

### 业务逻辑差异

**admserver biz**:
- Create: 允许设置任意状态
- Update: 可修改所有字段和状态
- List: 支持按任意状态筛选
- Delete: 软删除

**apiserver biz**:
- List: 强制过滤 `status=active`，忽略 status 参数
- Get: 只返回 active 状态的角色，否则返回 404

### 权限配置

在 `internal/pkg/database/seeder/api_seeder.go` 添加 AI Role API 记录：

```go
{
    Path:   "/v1/ai/roles",
    Method: "GET",
    Group:  "AI",
},
{
    Path:   "/v1/ai/roles",
    Method: "POST",
    Group:  "AI",
},
{
    Path:   "/v1/ai/roles/:id",
    Method: "GET",
    Group:  "AI",
},
{
    Path:   "/v1/ai/roles/:id",
    Method: "PUT",
    Group:  "AI",
},
{
    Path:   "/v1/ai/roles/:id",
    Method: "DELETE",
    Group:  "AI",
},
```

后台管理员通过 admserver 界面配置角色权限。

## 迁移步骤

### 阶段 1：准备 admserver

1. 创建 `internal/admserver/biz/ai/role.go`
2. 创建 `internal/admserver/handler/http/ai/role.go`
3. 在 `internal/admserver/router/api.go` 注册路由
4. 添加 Swagger 注释

### 阶段 2：简化 apiserver

1. 修改 `internal/apiserver/biz/chat/role.go`:
   - List 强制过滤 `status=active`
   - Get 只返回 active 角色
2. 简化 `internal/apiserver/handler/http/chat/role.go`:
   - 移除 Create/Update/Delete 方法
   - 更新 List/Get 的 Swagger 注释
3. 修改 `internal/apiserver/router/ai.go`:
   - 移除 POST/PUT/DELETE 路由

### 阶段 3：数据配置

1. 创建/更新 API seeder
2. 执行 `bingo db seed`

### 阶段 4：测试

1. 单元测试
2. 集成测试
3. 权限验证测试

## 测试策略

### 单元测试

**admserver**:
- Create: 创建不同状态的角色
- Update: 修改角色状态和字段
- Delete: 软删除
- List: 按状态筛选

**apiserver**:
- List: 只返回 active 角色
- Get: 只返回 active 角色，inactive 返回 404
- 参数验证：忽略 status 参数

### 集成测试

- 普通用户调用 apiserver 查询成功
- 普通用户尝试访问 admserver 接口返回 401/403
- 管理员通过 admserver 完整 CRUD
- 前端功能不受影响

## 风险和注意事项

1. **前端兼容性**: 确保前端调用的是 `/v1/ai/roles` GET 接口，不受影响
2. **数据迁移**: 无需迁移，使用现有表结构
3. **回滚方案**: 保留 git 历史，可快速回滚
4. **测试覆盖**: 确保 TDD 流程，测试先行

## 后续优化

1. 添加 AI Role 审核流程（草稿 → 审核中 → 已发布）
2. 添加使用统计和热门角色
3. 支持用户自定义角色（user_id 关联）
4. 添加角色版本管理
