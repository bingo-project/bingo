# 项目规范

> **本文件是所有代码规范的唯一来源。**
>
> AI 生成代码前必须读取此文件。其他文档提供详细说明，但规则以本文件为准。

---

## 核心原则：三层架构

所有业务代码必须遵循 Handler → Biz → Store 三层架构，**严禁跨层调用**。

```
┌─────────────────────────────────────────┐
│         Handler Layer                   │  HTTP/WebSocket/gRPC 处理层
│  - 参数验证、请求响应转换、错误处理      │
└──────────────┬──────────────────────────┘
               │ Depends on
               ▼
┌─────────────────────────────────────────┐
│          Business Layer (Biz)           │  业务逻辑层
│  - 业务规则、流程编排、事务控制          │
└──────────────┬──────────────────────────┘
               │ Depends on
               ▼
┌─────────────────────────────────────────┐
│          Store Layer                    │  数据访问层
│  - 数据库操作、缓存操作                  │
└─────────────────────────────────────────┘
```

**每层职责：**

| 层级 | 职责 | 禁止 |
|------|------|------|
| Handler | 参数绑定、调用 Biz、返回响应 | 包含业务逻辑 |
| Biz | 业务规则、流程编排、事务控制 | 直接操作数据库 |
| Store | 数据库 CRUD、缓存操作 | 包含业务逻辑 |

---

## 目录

1. [文件规范](#1-文件规范)
2. [命名规范](#2-命名规范)
3. [代码组织](#3-代码组织)
4. [错误处理](#4-错误处理)
5. [日志规范](#5-日志规范)
6. [测试规范](#6-测试规范)
7. [数据库迁移与初始化](#7-数据库迁移与初始化)
8. [生成代码检查清单](#8-生成代码检查清单)
9. [构建规范](#9-构建规范)
10. [API 规范](#10-api-规范)

---

## 1. 文件规范

### 1.1 ABOUTME 注释（必须）

每个 `.go` 文件必须以 2 行 ABOUTME 注释开头。

```go
// ❌ 禁止：没有 ABOUTME 注释
package user

// ❌ 禁止：只有 1 行
// ABOUTME: User business logic
package user

// ✅ 必须：2 行 ABOUTME 注释
// ABOUTME: User business logic implementation.
// ABOUTME: Handles user registration, login, and profile management.
package user
```

### 1.2 目录结构

```
internal/<server>/
├── app.go                  # 应用初始化
├── run.go                  # 服务启动逻辑
├── biz/                    # 业务逻辑层
│   ├── biz.go              # IBiz 接口定义
│   ├── auth/               # 认证业务
│   └── user/               # 用户业务
├── handler/                # Handler 层（支持多协议）
│   ├── http/               # HTTP Handler
│   │   ├── auth/
│   │   └── user/
│   ├── ws/                 # WebSocket Handler
│   │   └── auth.go
│   └── grpc/               # gRPC Handler
│       └── auth.go
└── router/                 # 路由定义

internal/pkg/
├── store/                  # 数据访问层（平铺，避免循环引用）
│   ├── store.go            # IStore 接口
│   ├── user.go             # 用户 Store
│   └── sys_config.go       # 系统配置 Store
├── model/                  # 数据模型
├── errno/                  # 错误码定义
└── auth/                   # 认证授权
```

---

## 2. 命名规范

### 2.1 包名

- 小写、简短、有意义
- 单数形式
- 不使用下划线或驼峰

```go
// ✅ 正确
package user
package auth

// ❌ 错误
package users          // 应该用单数
package userService    // 不使用驼峰
```

### 2.2 文件名

- 蛇形命名（snake_case）

```
user_handler.go
sys_config.go
auth_middleware.go
```

### 2.3 表名

- 统一使用**单数形式**
- 使用**模块前缀**，避免命名冲突
- 蛇形命名（snake_case）

| 模块 | 前缀 | 示例 |
|------|------|------|
| 系统 | `sys_` | `sys_config`, `sys_menu` |
| 通知 | `ntf_` | `ntf_message`, `ntf_announcement` |
| 用户 | `user_` 或无前缀 | `user`, `user_address` |

```sql
-- ✅ 正确
CREATE TABLE ntf_message (...);
CREATE TABLE sys_config (...);

-- ❌ 错误
CREATE TABLE notifications (...);  -- 应该用单数 + 前缀
CREATE TABLE message (...);        -- 缺少模块前缀
```

### 2.4 API JSON 字段命名

- **必须使用驼峰命名（camelCase）**
- **禁止使用蛇形命名（snake_case）**

```go
// ✅ 正确：驼峰命名
type UserInfo struct {
    UserID      string    `json:"userId"`
    Username    string    `json:"username"`
    Nickname    string    `json:"nickname"`
    CountryCode string    `json:"countryCode"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

// ❌ 错误：蛇形命名
type ChatRequest struct {
    SessionID string `json:"session_id"`  // 应该是 sessionId
    MaxTokens int    `json:"max_tokens"`  // 应该是 maxTokens
    RoleID    string `json:"role_id"`     // 应该是 roleId
}
```

**例外**: OpenAI 兼容接口保持原有 snake_case 命名以符合标准

```go
// ✅ OpenAI 兼容接口可以使用 snake_case
type OpenAIRequest struct {
    Model       string        `json:"model"`
    Messages    []ChatMessage `json:"messages"`
    MaxTokens   int           `json:"max_tokens"`   // OpenAI 标准
    Temperature float64       `json:"temperature"`  // OpenAI 标准
}
```

### 2.5 接口名

- `I` 前缀（Interface）
- 大驼峰命名

```go
type IStore interface {}
type IBiz interface {}
```

### 2.6 Store 命名规范

| 元素 | 规范 | 示例 |
|------|------|------|
| 文件名 | `<prefix>_<model>.go` | `user.go`, `sys_config.go` |
| Store 接口 | `<Model>Store` | `UserStore`, `SysConfigStore` |
| 实现结构体 | `<model>Store` (小写) | `userStore`, `sysConfigStore` |
| 扩展接口 | `<Model>Expansion` | `UserExpansion` |
| 创建函数 | `New<Model>Store()` | `NewUserStore()` |

---

## 3. 代码组织

### 3.1 HTTP Handler 层模板

**每个 HTTP Handler 方法必须编写 Swagger 注释**，用于生成 API 文档：

```go
// ABOUTME: HTTP handlers for user management.
// ABOUTME: Provides CRUD endpoints for user resources.
package user

type UserHandler struct {
    biz biz.IBiz
}

func New(biz biz.IBiz) *UserHandler {
    return &UserHandler{biz: biz}
}

// Create
// @Summary    Create user
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      request  body      v1.CreateUserRequest  true  "Param"
// @Success    200      {object}  v1.UserInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/users [POST].
func (h *UserHandler) Create(c *gin.Context) {
    // 1. 参数验证
    var req v1.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
        return
    }

    // 2. 调用业务层（直接传 c，不用 c.Request.Context()）
    user, err := h.biz.Users().Create(c, &req)

    // 3. 返回响应
    core.Response(c, user, err)
}
```

**Swagger 注释说明：**

| 注解 | 说明 | 示例 |
|------|------|------|
| `@Summary` | 接口简要描述 | `Create user` |
| `@Security` | 认证方式（需要登录则加） | `Bearer` |
| `@Tags` | 接口分组 | `User`, `Auth` |
| `@Accept` | 请求格式 | `application/json` |
| `@Produce` | 响应格式 | `application/json` |
| `@Param` | 参数定义 | `request body v1.CreateUserRequest true "Param"` |
| `@Success` | 成功响应 | `200 {object} v1.UserInfo` |
| `@Failure` | 错误响应 | `400 {object} core.ErrResponse` |
| `@Router` | 路由路径和方法 | `/v1/users [POST]` |

**注意**：`@Router` 注解末尾需要加 `.` 以符合 golint 规范。

### 3.2 WebSocket Handler 层模板

WebSocket 使用 JSON-RPC 2.0 协议，返回 `*jsonrpc.Response`。

```go
// ABOUTME: WebSocket auth method handlers.
// ABOUTME: Provides login and user-info endpoints for WS clients.
package ws

import (
    "github.com/bingo-project/websocket"
    "github.com/bingo-project/websocket/jsonrpc"

    "<module>/internal/<server>/biz"
    "<module>/internal/pkg/errno"
    v1 "<module>/pkg/api/<server>/v1"
)

type Handler struct {
    b biz.IBiz
}

func NewHandler(ds store.IStore) *Handler {
    return &Handler{b: biz.NewBiz(ds)}
}

func (h *Handler) Login(c *websocket.Context) *jsonrpc.Response {
    // 1. 参数绑定和验证
    var req v1.LoginRequest
    if err := c.BindValidate(&req); err != nil {
        return c.Error(errno.ErrInvalidArgument.WithMessage("%s", err.Error()))
    }

    // 2. 调用业务层
    resp, err := h.b.Auth().Login(c, &req)
    if err != nil {
        return c.Error(err)
    }

    // 3. 返回 JSON-RPC 响应
    return c.JSON(resp)
}
```

### 3.3 gRPC Handler 层模板

gRPC 使用 Protobuf，需嵌入 `Unimplemented*Server`。

```go
// ABOUTME: gRPC auth method handlers.
// ABOUTME: Provides login and user-info endpoints for gRPC clients.
package grpc

import (
    "context"

    "<module>/internal/<server>/biz"
    "<module>/internal/pkg/errno"
    apiv1 "<module>/pkg/api/<server>/v1"
    v1 "<module>/pkg/proto/<server>/v1/pb"
)

type Handler struct {
    b biz.IBiz
    v1.UnimplementedApiServerServer  // 必须嵌入
}

func NewHandler(ds store.IStore) *Handler {
    return &Handler{b: biz.NewBiz(ds)}
}

func (h *Handler) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
    // 1. 参数转换和验证
    loginReq := &apiv1.LoginRequest{
        Account:  req.Account,
        Password: req.Password,
    }
    if err := validate.Struct(loginReq); err != nil {
        return nil, errno.ErrInvalidArgument
    }

    // 2. 调用业务层
    resp, err := h.b.Auth().Login(ctx, loginReq)
    if err != nil {
        return nil, err
    }

    // 3. 返回 Protobuf 响应
    return &v1.LoginReply{
        AccessToken: resp.AccessToken,
        TokenType:   "Bearer",
        ExpiresIn:   resp.ExpiresAt.Unix(),
    }, nil
}
```

### 3.4 Handler 协议对比

| 特性 | HTTP | WebSocket | gRPC |
|------|------|-----------|------|
| Context | `*gin.Context` | `*websocket.Context` | `context.Context` |
| 返回值 | `void` | `*jsonrpc.Response` | `(*Reply, error)` |
| 参数绑定 | `c.ShouldBindJSON()` | `c.BindValidate()` | 直接使用 Protobuf |
| 成功响应 | `core.Response(c, data, nil)` | `return c.JSON(data)` | `return &Reply{}, nil` |
| 错误响应 | `core.Response(c, nil, err)` | `return c.Error(err)` | `return nil, err` |
| 协议格式 | RESTful JSON | JSON-RPC 2.0 | Protobuf |

### 3.5 Context 传递规范

**HTTP Handler 调用 Biz 层时，直接传 `c`，不需要 `c.Request.Context()`：**

```go
// ✅ 正确：直接传 c
uid := contextx.UserID(c)
user, err := h.biz.Users().Create(c, &req)

// ❌ 错误：不必要的 c.Request.Context()
uid := contextx.UserID(c.Request.Context())
user, err := h.biz.Users().Create(c.Request.Context(), &req)
```

`*gin.Context` 实现了 `context.Context` 接口，直接传递即可。

### 3.6 Biz 层模板

```go
// ABOUTME: User business logic implementation.
// ABOUTME: Handles user creation, validation, and password encryption.
package user

type UserBiz interface {
    Create(ctx context.Context, req *v1.CreateUserRequest) (*model.User, error)
}

type userBiz struct {
    ds store.IStore
}

func New(ds store.IStore) UserBiz {
    return &userBiz{ds: ds}
}

func (b *userBiz) Create(ctx context.Context, req *v1.CreateUserRequest) (*model.User, error) {
    // 1. 业务规则验证
    if err := b.validateUser(req); err != nil {
        return nil, err
    }

    // 2. 业务逻辑处理
    user := &model.User{
        Username: req.Username,
        Password: auth.Encrypt(req.Password),
    }

    // 3. 数据持久化
    if err := b.ds.Users().Create(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}
```

### 3.7 Store 层模板

```go
// ABOUTME: User data access layer.
// ABOUTME: Provides CRUD operations for user records.
package store

type UserStore interface {
    Create(ctx context.Context, obj *model.UserM) error
    Get(ctx context.Context, opts *where.Options) (*model.UserM, error)
    // ... 其他 CRUD 方法

    UserExpansion  // 扩展接口
}

type UserExpansion interface {
    FindByEmail(ctx context.Context, email string) (*model.UserM, error)
}

type userStore struct {
    *genericstore.Store[model.UserM]
}

func NewUserStore(store *datastore) *userStore {
    return &userStore{
        Store: genericstore.NewStore[model.UserM](store, NewLogger()),
    }
}

func (s *userStore) FindByEmail(ctx context.Context, email string) (*model.UserM, error) {
    return s.Get(ctx, where.F("email", email))
}
```

### 3.8 响应数据结构规范

#### 3.8.1 数据拼装位置

**响应数据结构由 Biz 层组装，Handler 层直接返回 Biz 层的响应：**

```go
// ✅ 正确：Biz 层组装响应
func (b *userBiz) List(ctx context.Context, req *v1.ListUserRequest) (*v1.ListUserResponse, error) {
    users, total, err := b.ds.Users().List(ctx, opts)
    return &v1.ListUserResponse{
        Total: total,
        Data:  users,
    }, err
}

// Handler 层直接返回
func (h *UserHandler) List(c *gin.Context) {
    resp, err := h.biz.Users().List(c, &req)
    core.Response(c, resp, err)
}

// ❌ 错误：Handler 层组装数据结构
func (h *UserHandler) List(c *gin.Context) {
    users, total, err := h.biz.Users().List(c, &req)
    core.Response(c, &v1.ListUserResponse{
        Total: total,
        Data:  users,
    }, err)  // 禁止！拼装应在 Biz 层
}
```

#### 3.8.2 分页列表查询

分页列表查询必须返回包含 `Total` 和 `Data` 字段的结构体：

```go
// ✅ 正确：Biz 层返回分页响应结构体
type ListUserResponse struct {
    Total int64      `json:"total"`
    Data  []UserInfo `json:"data"`
}

// @Success 200 {object} v1.ListUserResponse
```

#### 3.8.3 非分页列表查询

非分页列表查询**直接返回切片**，**禁止嵌套 `data` 字段**：

```go
// ❌ 错误：非分页列表不应嵌套 data
type ListSessionsResponse struct {
    Data []SessionInfo `json:"data"`
}

// ✅ 正确：Biz 层直接返回切片
func (b *sessionBiz) List(ctx context.Context) ([]SessionInfo, error) {
    sessions, err := b.ds.Sessions().List(ctx)
    return sessions, err
}

// Handler 层直接返回
func (h *SessionHandler) List(c *gin.Context) {
    sessions, err := h.biz.Sessions().List(c)
    core.Response(c, sessions, err)
}

// ✅ Swagger 注解
// @Success 200 {object} []v1.SessionInfo
```

#### 3.8.4 单对象查询

单对象查询直接返回对象结构体：

```go
// ✅ 正确：Biz 层直接返回对象
func (b *userBiz) Get(ctx context.Context, uid string) (*UserInfo, error) {
    user, err := b.ds.Users().Get(ctx, uid)
    return user, err
}

// ✅ Swagger 注解
// @Success 200 {object} v1.UserInfo
```

#### 3.8.5 判断准则

| 场景 | 响应类型 | Biz 层返回类型 | Handler 调用 |
|------|----------|----------------|--------------|
| 分页列表 | 带分页信息的结构体 | `*ListXxxResponse` | `core.Response(c, resp, err)` |
| 非分页列表 | 直接返回切片 | `[]XxxInfo` | `core.Response(c, data, err)` |
| 单对象 | 直接返回结构体 | `*XxxInfo` | `core.Response(c, data, err)` |
| 创建/更新 | 返回创建后的对象 | `*XxxInfo` | `core.Response(c, data, err)` |
| 删除 | 返回 nil | `error` | `core.Response(c, nil, err)` |

---

## 4. 错误处理

### 4.1 统一错误码

错误码定义在 `internal/pkg/errno/`：

```go
var (
    ErrUserNotFound      = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.UserNotFound", Message: "User was not found."}
    ErrUserAlreadyExist  = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.UserAlreadyExist", Message: "User already exist."}
    ErrPasswordInvalid   = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.PasswordInvalid", Message: "Password is incorrect."}
)
```

### 4.2 分层错误处理

| 层级 | 错误处理方式 |
|------|-------------|
| Store 层 | 返回原始 error（GORM/Redis 错误） |
| Biz 层 | **必须**转换为自定义错误码，用 `WithMessage` 附加上下文 |
| Handler 层 | 直接传递给 `core.Response()` |

**重要**：Biz 层**禁止**直接返回 Store 层的错误，必须使用 `errno` 包装。

```go
// Store 层 - 返回原始 error
func (s *userStore) Create(ctx context.Context, user *model.UserM) error {
    return s.db.Create(user).Error
}

// Biz 层 - 转换为自定义错误码
func (b *userBiz) Create(ctx context.Context, req *v1.CreateUserRequest) (*model.UserM, error) {
    // ✅ 正确：业务规则校验返回自定义错误码
    if exists, _ := b.ds.Users().FindByEmail(ctx, req.Email); exists != nil {
        return nil, errno.ErrUserAlreadyExist
    }

    user := &model.UserM{Email: req.Email}
    if err := b.ds.Users().Create(ctx, user); err != nil {
        // ✅ 正确：写操作用 ErrDBWrite，读操作用 ErrDBRead
        return nil, errno.ErrDBWrite.WithMessage("create user: %v", err)
    }

    return user, nil
}

// ❌ 错误：直接返回 Store 层的 error
if err := b.ds.Users().Create(ctx, user); err != nil {
    return nil, err  // 禁止！必须用 errno 包装
}

// ❌ 错误：用 fmt.Errorf 包装（丢失类型信息）
return nil, fmt.Errorf("failed to create user: %w", err)

// ❌ 错误：直接返回字符串错误
return nil, errors.New("用户不存在")
```

**常用错误码：**

| 场景 | 错误码 |
|------|--------|
| 数据库读操作失败 | `errno.ErrDBRead` |
| 数据库写操作失败 | `errno.ErrDBWrite` |
| 资源不存在 | `errno.ErrNotFound` |
| 权限不足 | `errno.ErrPermissionDenied` |
| 操作失败（如队列入队） | `errno.ErrOperationFailed` |

### 4.3 Handler 层统一响应

```go
// ✅ 必须使用 core.Response
core.Response(c, data, err)

// ❌ 禁止直接返回 JSON
c.JSON(200, gin.H{"data": data})
```

### 4.4 使用常量代替字面量

状态值、类型值等必须定义为常量，禁止在代码中直接使用字符串字面量。

**常量定义位置：**

| 常量类型 | 定义位置 | 示例 |
|---------|---------|------|
| Model 状态/类型 | `internal/pkg/model/` | `model.AdminStatus`, `model.GoogleStatus` |
| 错误码 | `internal/pkg/errno/` | `errno.ErrNotFound` |
| 已知值（如角色名） | `internal/pkg/known/` | `known.UserRoot` |
| 业务层专用常量 | 对应 biz 包 | `TOTPIssuer` |

```go
// ✅ 正确：在 model 包定义类型和常量
// internal/pkg/model/admin.go
type GoogleStatus string

const (
    GoogleStatusUnbind   GoogleStatus = "unbind"
    GoogleStatusDisabled GoogleStatus = "disabled"
    GoogleStatusEnabled  GoogleStatus = "enabled"
)

// ✅ 正确：在 biz 层使用 model 常量
if user.GoogleStatus == string(model.GoogleStatusEnabled) {
    // ...
}

// ❌ 错误：直接使用字符串字面量
if user.GoogleStatus == "enabled" {
    // ...
}

// ❌ 错误：在 biz 层重复定义常量
const GoogleStatusEnabled = "enabled"  // 应该使用 model.GoogleStatusEnabled
```

---

## 5. 日志规范

### 5.1 使用结构化日志

```go
import "github.com/bingo-project/bingo/internal/pkg/log"

// ✅ 结构化日志
log.C(ctx).Infow("user created", "username", username, "user_id", userID)

// ✅ 错误日志（带上下文）
log.C(ctx).Errorw("failed to create user", "err", err, "username", username)

// ❌ 不推荐：非结构化日志
log.C(ctx).Info("user created: " + username)
```

### 5.2 日志级别

- **Debug**: 调试信息
- **Info**: 重要业务流程
- **Warn**: 警告信息，不影响主流程
- **Error**: 错误信息，需要关注

---

## 6. 测试规范

### 6.1 分层测试策略

| 层级 | 测试方式 | Mock 什么 |
|------|----------|-----------|
| Store 层 | SQLite 内存数据库 | 不 mock |
| Biz 层 | Mock Store | Mock `store.IStore` |
| Handler 层 | Mock Biz | Mock `biz.IBiz` |

### 6.2 测试文件命名

```
user.go       -> user_test.go
article.go    -> article_test.go
```

### 6.3 测试用例模板

```go
func TestUserBiz_Create(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        // 1. 准备测试数据
        store := mockstore.NewStore()
        biz := user.New(store)
        req := &CreateUserRequest{Username: "test"}

        // 2. 执行测试
        user, err := biz.Create(context.Background(), req)

        // 3. 断言
        require.NoError(t, err)
        assert.Equal(t, "test", user.Username)
    })

    t.Run("user_already_exists", func(t *testing.T) {
        // ... 测试异常路径
    })
}
```

### 6.4 Mock 代码组织

```
internal/pkg/testing/mock/
├── store/           # Store 层 mock
│   └── store.go
└── biz/             # Biz 层 mock
    └── biz.go
```

---

## 7. 数据库迁移与初始化

### 7.1 目录结构

```
internal/pkg/database/
├── migration/          # 数据库迁移文件
└── seeder/             # 数据初始化文件
```

### 7.2 Migration（数据库迁移）

新增或变更表结构时，必须编写 migration 文件。

> **注意**：`bingo` 是全局安装的 CLI 工具，不是项目内的命令。

```bash
bingo migrate up        # 执行迁移
bingo migrate rollback  # 回滚上一次迁移
bingo migrate reset     # 重置所有迁移（⚠️ 仅开发环境）
```

#### 7.2.1 数据库字段规范

**基本原则**：

1. **仅允许必要的字段为 NULL**
   - 可选字段（如 `description`, `icon`）可为 NULL
   - 有明确业务含义的字段应设置 NOT NULL

2. **使用默认值代替 NULL**
   - 数值类型：使用 `0` 作为默认值（`default:0`）
   - 字符串类型：使用空字符串 `''` 作为默认值（`default:''`）
   - 布尔类型：使用 `false` 作为默认值
   - 时间戳：使用 `CURRENT_TIMESTAMP` 作为默认值

3. **有默认值的字段必须 NOT NULL**
   - 如果字段有 `default` 约束，必须同时声明 `not null`
   - 这确保默认值始终生效，避免 NULL 值

**示例**：

```go
// ✅ 正确：可选字段允许 NULL
Description string `gorm:"type:varchar(255)"`                          // 可选描述
Icon        string `gorm:"type:varchar(255)"`                          // 可选图标

// ✅ 正确：有默认值的字段必须 NOT NULL
Status      string `gorm:"type:varchar(16);not null;default:'active'"`
Sort        int    `gorm:"type:int;not null;default:0"`
Temperature float64 `gorm:"type:decimal(3,2);not null;default:0.70"`

// ✅ 正确：必填字段 NOT NULL
Name        string `gorm:"type:varchar(64);not null"`
Email       string `gorm:"type:varchar(128);not null"`

// ❌ 错误：有默认值但允许 NULL
Temperature float64 `gorm:"type:decimal(3,2);default:0.70"`  // 缺少 not null
Sort        int    `gorm:"type:int;default:0"`               // 缺少 not null

// ❌ 错误：能用默认值的不要用 NULL
Count       int    `gorm:"type:int"`                          // 应该 default:0
Enabled     bool   `gorm:"type:bool"`                         // 应该 default:false
```

**外键字段**：

```go
// ✅ 正确：可选关联允许 NULL
RoleID      string `gorm:"type:varchar(64);index:idx_role_id"`
ParentID    *uint  `gorm:"index:idx_parent_id"`  // 指针类型表示可空

// ✅ 正确：必填关联 NOT NULL
UserID      string `gorm:"type:varchar(64);not null;index:idx_user_id"`
```

**时间字段**：

```go
// ✅ 正确：创建和更新时间必须有默认值
CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`

// ✅ 正确：可选时间允许 NULL（使用指针）
DeletedAt *time.Time `gorm:"type:DATETIME(3)"`
LastLogin *time.Time `gorm:"type:DATETIME(3)"`
```


### 7.3 Seeder（数据初始化）

```bash
bingo db seed                       # 执行所有 seeder
bingo db seed --seeder=UserSeeder   # 执行指定 seeder
```

### 7.4 开发阶段 Migration 管理

功能分支合并前，可将多次表变更合并为单个 migration：

1. 新增 alter table migration 时，同步更新原 create table migration
2. 执行 `bingo migrate reset && bingo migrate up && bingo db seed` 验证
3. 删除临时的 alter table migration

### 7.5 测试数据 Seeder

维护 `UserSeeder` 等测试数据 seeder，便于 reset 后快速恢复开发环境。

---

## 8. 生成代码检查清单

**生成任何代码前，必须逐条确认：**

### 架构规范

- [ ] 遵循三层架构（Handler → Biz → Store）
- [ ] 无跨层调用（Handler 不直接调用 Store）
- [ ] 业务逻辑在 Biz 层，不在 Handler 或 Store 层

### 文件规范

- [ ] 文件头有 2 行 ABOUTME 注释
- [ ] 文件放在正确的目录
- [ ] 命名符合规范

### 错误处理

- [ ] 使用定义的错误码（`errno.ErrXxx`）
- [ ] Handler 层使用 `core.Response()` 返回
- [ ] 错误有适当的上下文信息

### 日志规范

- [ ] 使用结构化日志 `log.C(ctx).Infow()`
- [ ] 关键业务操作有日志记录
- [ ] 不记录敏感信息（密码等）

### 测试规范

- [ ] 每层只测自己的职责
- [ ] Store 层使用 SQLite 测试
- [ ] Biz/Handler 层使用 Mock
- [ ] 测试覆盖正常和异常路径

### 数据库变更

- [ ] 表结构变更有对应的 migration 文件
- [ ] migration 支持 rollback
- [ ] 必要的初始化数据有 seeder

---

## 9. 构建规范

### 9.1 代码修改后重新构建

代码修改后需要重新构建，**使用 `make build`，不要用 `go build ./...`**：

```bash
# ✅ 正确
make build

# ❌ 错误
go build ./...
```

### 9.2 修改 API 参数定义

如果修改了请求/响应结构体（`pkg/api/` 下的定义），必须**先执行 `make swag` 再构建**：

```bash
# 修改了 v1.CreateUserRequest 等结构体后
make swag   # 先更新 Swagger 文档
make build  # 再构建
```

### 9.3 提交前检查

**commit 前必须执行 `make lint`**，确保代码符合规范：

```bash
make lint   # 代码检查
git add .
git commit -m "feat: add feature"
```

---

## 附录：常用命令

```bash
make build                           # 编译所有服务
make build BINS="svc1 svc2"          # 编译指定服务（可多个）
make test                            # 测试
make lint                            # 代码检查（commit 前必须执行）
make swag                            # 生成 Swagger 文档（仅 HTTP 服务）
```

## 附录：Import 路径

```go
// Gin 框架
import "github.com/gin-gonic/gin"

// 业务层
import "<module>/internal/<server>/biz"

// 数据层
import "<module>/internal/pkg/store"

// 错误码
import "<module>/internal/pkg/errno"

// 日志
import "<module>/internal/pkg/log"
log.C(ctx).Infow("message", "key", value)

// 统一响应
import "<module>/internal/pkg/core"
core.Response(c, data, err)
```

---

## 10. API 规范

### 10.1 JSON 字段命名

**基本原则:**
- **必须使用驼峰命名（camelCase）**
- **禁止使用蛇形命名（snake_case）**
- **唯一例外**: OpenAI 兼容接口可使用 snake_case 以符合标准

```go
// ✅ 正确：驼峰命名
type UserInfo struct {
    UserID      string    `json:"userId"`
    Username    string    `json:"username"`
    Nickname    string    `json:"nickname"`
    CountryCode string    `json:"countryCode"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

// ✅ 正确：OpenAI 兼容接口
type OpenAIRequest struct {
    Model       string        `json:"model"`
    Messages    []ChatMessage `json:"messages"`
    MaxTokens   int           `json:"max_tokens"`   // OpenAI 标准
    Temperature float64       `json:"temperature"`  // OpenAI 标准
}

// ❌ 错误：非 OpenAI 接口使用 snake_case
type ChatRequest struct {
    SessionID string `json:"session_id"`  // 应该是 sessionId
    MaxTokens int    `json:"max_tokens"`  // 应该是 maxTokens
    RoleID    string `json:"role_id"`     // 应该是 roleId
}
```

### 10.2 Request/Response 结构

**命名规范:**
- 请求结构: `<Verb><Resource>Request`
- 响应结构: `<Verb><Resource>Response` 或 `<Resource>Info`
- 列表响应: `List<Resource>Response`

```go
// 请求
type CreateUserRequest struct {}
type UpdateUserRequest struct {}
type DeleteUserRequest struct {}

// 响应
type UserInfo struct {}
type ListUserResponse struct {
    Total int64       `json:"total"`
    Data  []UserInfo  `json:"data"`
}
```

**字段顺序:**
1. 资源标识字段 (ID, UID 等)
2. 核心业务字段
3. 状态字段
4. 时间字段

```go
type UserInfo struct {
    UID          string    `json:"uid"`           // 1. 标识
    Username     string    `json:"username"`      // 2. 业务字段
    Nickname     string    `json:"nickname"`
    Status       int32     `json:"status"`        // 3. 状态
    CreatedAt    time.Time `json:"createdAt"`     // 4. 时间
    UpdatedAt    time.Time `json:"updatedAt"`
}
```

### 10.3 HTTP 状态码使用

| 场景 | 状态码 | 说明 |
|------|--------|------|
| 成功 | 200 | 统一使用 200，通过业务错误码区分具体状态 |
| 参数错误 | 200 | 返回 errno.ErrInvalidArgument |
| 未授权 | 200 | 返回 errno.ErrUnauthorized |
| 资源不存在 | 200 | 返回 errno.ErrNotFound |
| 服务器错误 | 200 | 返回 errno.ErrInternal |

**原则**: HTTP 层始终返回 200，业务错误通过 `errno` 和 `core.Response` 处理


> `<module>` = go.mod 模块名，`<server>` = 服务名（如 apiserver）
