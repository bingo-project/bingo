# 统一认证授权

Bingo 提供了插件式的认证授权架构，通用逻辑集中在 `internal/pkg/auth`，各服务通过实现接口来定制用户加载和权限解析逻辑。

## 特性

- **插件式设计**：通过 `UserLoader` 和 `SubjectResolver` 接口支持不同用户类型
- **协议统一**：HTTP 和 gRPC 共享相同的认证授权逻辑
- **职责分离**：认证（authn）和授权（authz）独立，可单独使用
- **类型安全**：context 携带类型化的用户信息

## 目录结构

```
internal/pkg/auth/
├── authenticator.go      # 认证器核心（token 验证 + UserLoader 调用）
├── authenticator_test.go # 认证器测试
├── authorizer.go         # 授权器核心（Casbin 封装 + SubjectResolver 调用）
├── interceptor.go        # gRPC interceptors
├── middleware.go         # HTTP middleware
├── password.go           # 密码加密/比较
└── websocket.go          # WebSocket helpers

internal/apiserver/biz/auth/
└── loader.go             # apiserver UserLoader 实现

internal/admserver/biz/auth/
└── loader.go             # admserver AdminLoader + SubjectResolver 实现
```

## 核心接口

### UserLoader - 加载用户信息

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

### SubjectResolver - 获取授权主体

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

## 服务实现示例

### apiserver UserLoader

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

### admserver AdminLoader + SubjectResolver

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

## 使用方式

### apiserver HTTP (router/api.go)

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

### apiserver gRPC (grpc.go)

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

### admserver HTTP (router/api.go)

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

## 相关文档

- [可插拔协议层](protocol-layer.md) - HTTP/gRPC/WebSocket 统一架构
- [统一错误处理](unified-error-handling.md) - 三协议共享错误格式
