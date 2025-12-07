# 统一认证设计

## 问题

当前认证和授权逻辑分散在多处：

**认证 (authn)：**
- `internal/pkg/auth` - 只做 token 验证，不加载用户
- `internal/apiserver/middleware/authn.go` - HTTP 认证，查 User 表
- `internal/admserver/middleware/authn.go` - HTTP 认证，查 Admin 表
- `internal/pkg/middleware/grpc/authn.go` - gRPC 认证，查 User 表
- `pkg/ws/middleware/auth.go` - WS 认证，依赖 Client 状态

**授权 (authz)：**
- `internal/admserver/middleware/authz.go` - HTTP 授权，admserver 专用
- `pkg/auth/authz.go` - Casbin 封装

路径分散，职责不清晰，代码重复。

## 目标

1. 统一认证入口到 `internal/pkg/auth`
2. 统一授权入口到 `internal/pkg/auth`
3. 支持不同用户类型（user, admin 等）
4. 保持通用模块不耦合具体 model

## 设计

### 目录结构

```
internal/pkg/auth/
├── authenticator.go      # 认证器核心（token 验证 + UserLoader 调用）
├── authorizer.go         # 授权器核心（Casbin 封装 + SubjectResolver 调用）
├── interceptor.go        # gRPC interceptors
├── middleware.go         # HTTP middleware
├── websocket.go          # WebSocket helpers
└── authenticator_test.go
```

### 核心接口

#### UserLoader - 加载用户信息

```go
// internal/pkg/auth/authenticator.go

// UserLoader 加载用户信息到 context
type UserLoader interface {
    LoadUser(ctx context.Context, userID string) (context.Context, error)
}

// Authenticator 统一认证器
type Authenticator struct {
    loader UserLoader
}

// New 创建认证器，loader 可选
func New(loader UserLoader) *Authenticator {
    return &Authenticator{loader: loader}
}

// Verify 验证 token 并加载用户信息
func (a *Authenticator) Verify(ctx context.Context, tokenStr string) (context.Context, error) {
    if tokenStr == "" {
        return ctx, errorsx.New(401, "Unauthenticated", "token is required")
    }

    payload, err := token.Parse(tokenStr)
    if err != nil {
        return ctx, errorsx.New(401, "Unauthenticated", "invalid token: %s", err.Error())
    }

    ctx = contextx.WithUserID(ctx, payload.Subject)

    if a.loader != nil {
        ctx, err = a.loader.LoadUser(ctx, payload.Subject)
        if err != nil {
            return ctx, err
        }
    }

    return ctx, nil
}
```

#### SubjectResolver - 获取授权主体

```go
// internal/pkg/auth/authorizer.go

// SubjectResolver 从 context 获取授权主体
type SubjectResolver interface {
    ResolveSubject(ctx context.Context) (string, error)
}

// Authorizer 统一授权器
type Authorizer struct {
    enforcer *casbin.Enforcer
    resolver SubjectResolver
}

// NewAuthorizer 创建授权器
func NewAuthorizer(db *gorm.DB, resolver SubjectResolver) (*Authorizer, error) {
    // 初始化 casbin...
}

// Authorize 检查权限
func (a *Authorizer) Authorize(ctx context.Context, obj, act string) error {
    sub, err := a.resolver.ResolveSubject(ctx)
    if err != nil {
        return err
    }

    allowed, err := a.enforcer.Enforce(sub, obj, act)
    if err != nil {
        return err
    }
    if !allowed {
        return errorsx.New(403, "Forbidden", "permission denied")
    }

    return nil
}
```

### apiserver 实现

```go
// internal/apiserver/biz/auth/loader.go

type UserLoader struct {
    store store.IStore
}

func NewUserLoader(store store.IStore) *UserLoader {
    return &UserLoader{store: store}
}

func (l *UserLoader) LoadUser(ctx context.Context, userID string) (context.Context, error) {
    user, err := l.store.User().GetByUID(ctx, userID)
    if err != nil {
        return ctx, errorsx.New(401, "Unauthenticated", "user not found")
    }

    var userInfo v1.UserInfo
    _ = copier.Copy(&userInfo, user)
    userInfo.PayPassword = user.PayPassword != ""

    ctx = contextx.WithUserInfo(ctx, &userInfo)
    ctx = contextx.WithUsername(ctx, userInfo.Username)

    return ctx, nil
}
```

### admserver 实现

```go
// internal/admserver/biz/auth/loader.go

type AdminLoader struct {
    store store.IStore
}

func NewAdminLoader(store store.IStore) *AdminLoader {
    return &AdminLoader{store: store}
}

func (l *AdminLoader) LoadUser(ctx context.Context, userID string) (context.Context, error) {
    admin, err := l.store.Admin().GetUserInfo(ctx, userID)
    if err != nil {
        return ctx, errorsx.New(401, "Unauthenticated", "admin not found")
    }

    var adminInfo v1.AdminInfo
    _ = copier.Copy(&adminInfo, admin)

    ctx = contextx.WithUserInfo(ctx, &adminInfo)
    ctx = contextx.WithUsername(ctx, adminInfo.Username)

    return ctx, nil
}

// AdminSubjectResolver 从 admin 信息获取授权主体
type AdminSubjectResolver struct{}

func (r *AdminSubjectResolver) ResolveSubject(ctx context.Context) (string, error) {
    admin, ok := contextx.UserInfo[v1.AdminInfo](ctx)
    if !ok {
        return "", errorsx.New(401, "Unauthenticated", "admin info not found")
    }

    return known.RolePrefix + admin.RoleName, nil
}
```

### 使用方式

#### apiserver HTTP (router/api.go)

```go
func MapApiRouters(g *gin.Engine) {
    loader := apiauth.NewUserLoader(store.S)
    authn := auth.New(loader)

    v1 := g.Group("/v1")

    // 公开路由...

    v1.Use(auth.Middleware(authn))

    // 需要认证的路由...
}
```

#### apiserver gRPC (grpc.go)

```go
func initGRPCServer(cfg *config.GRPC) *grpc.Server {
    loader := apiauth.NewUserLoader(store.S)
    authn := auth.New(loader)

    opts := []grpc.ServerOption{
        grpc.ChainUnaryInterceptor(
            interceptor.RequestID,
            interceptor.ClientIP,
            interceptor.Logger,
            interceptor.Recovery,
            interceptor.Validator,
            auth.UnaryInterceptor(authn, publicMethods),
        ),
    }
    // ...
}

var publicMethods = map[string]bool{
    "/apiserver.v1.ApiServer/Healthz": true,
    "/apiserver.v1.ApiServer/Version": true,
    "/apiserver.v1.ApiServer/Login":   true,
}
```

#### admserver HTTP (router/api.go)

```go
func MapApiRouters(g *gin.Engine) {
    loader := admauth.NewAdminLoader(store.S)
    authn := auth.New(loader)

    resolver := &admauth.AdminSubjectResolver{}
    authz, _ := auth.NewAuthorizer(store.S.DB(ctx), resolver)

    v1 := g.Group("/v1")

    // 公开路由...

    v1.Use(auth.Middleware(authn))

    // 需要认证的路由...

    v1.Use(auth.AuthzMiddleware(authz))

    // 需要授权的路由...
}
```

## 文件变更

### 修改

- `internal/pkg/auth/authenticator.go` - 添加 UserLoader 接口
- `internal/pkg/auth/interceptor.go` - 支持公开方法白名单参数
- `internal/pkg/auth/middleware.go` - 简化，使用 Authenticator

### 新增

- `internal/pkg/auth/authorizer.go` - 授权器（从 pkg/auth/authz.go 迁移核心逻辑）
- `internal/apiserver/biz/auth/loader.go` - apiserver UserLoader
- `internal/admserver/biz/auth/loader.go` - admserver AdminLoader + SubjectResolver

### 删除

- `internal/apiserver/middleware/authn.go`
- `internal/admserver/middleware/authn.go`
- `internal/admserver/middleware/authz.go`
- `internal/pkg/middleware/grpc/authn.go`

### 保留

- `pkg/auth/authz.go` - 保留对外暴露的 Authz 类型（如果有外部依赖）
- `pkg/ws/middleware/auth.go` - WS 认证逻辑不同，保持独立

## 迁移步骤

1. 修改 `internal/pkg/auth/authenticator.go`，添加 UserLoader 接口
2. 修改 `internal/pkg/auth/interceptor.go`，支持白名单参数
3. 修改 `internal/pkg/auth/middleware.go`，使用 Authenticator
4. 创建 `internal/pkg/auth/authorizer.go`
5. 创建 `internal/apiserver/auth/loader.go`
6. 修改 apiserver 的 grpc.go 和 router/api.go
7. 删除 `internal/apiserver/middleware/authn.go`
8. 删除 `internal/pkg/middleware/grpc/authn.go`
9. 创建 `internal/admserver/auth/loader.go`
10. 修改 admserver 的 grpc.go 和 router/api.go
11. 删除 `internal/admserver/middleware/authn.go` 和 `authz.go`
12. 运行测试验证
