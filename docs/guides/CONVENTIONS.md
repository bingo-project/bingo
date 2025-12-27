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
7. [生成代码检查清单](#7-生成代码检查清单)

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

### 2.3 接口名

- `I` 前缀（Interface）
- 大驼峰命名

```go
type IStore interface {}
type IBiz interface {}
```

### 2.4 Store 命名规范

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

func (h *UserHandler) Create(c *gin.Context) {
    // 1. 参数验证
    var req v1.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
        return
    }

    // 2. 调用业务层
    user, err := h.biz.Users().Create(c.Request.Context(), &req)

    // 3. 返回响应
    core.Response(c, user, err)
}
```

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

### 3.5 Biz 层模板

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

### 3.6 Store 层模板

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
| Biz 层 | 转换为自定义错误码，用 `WithMessage` 附加上下文 |
| Handler 层 | 直接传递给 `core.Response()` |

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
        // ✅ 正确：用 WithMessage 附加上下文
        return nil, errno.ErrDatabase.WithMessage("create user failed: %v", err)
    }

    return user, nil
}

// ❌ 错误：用 fmt.Errorf 包装（丢失类型信息）
return nil, fmt.Errorf("failed to create user: %w", err)

// ❌ 错误：直接返回字符串错误
return nil, errors.New("用户不存在")
```

### 4.3 Handler 层统一响应

```go
// ✅ 必须使用 core.Response
core.Response(c, data, err)

// ❌ 禁止直接返回 JSON
c.JSON(200, gin.H{"data": data})
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

## 7. 生成代码检查清单

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

> `<module>` = go.mod 模块名，`<server>` = 服务名（如 apiserver）
