# Unified Authentication and Authorization

Bingo provides a plugin-based authentication and authorization architecture. Common logic is centralized in `internal/pkg/auth`, while services implement interfaces to customize user loading and permission resolution.

## Features

- **Plugin-based Design**: Support different user types via `UserLoader` and `SubjectResolver` interfaces
- **Protocol Unified**: HTTP and gRPC share the same authentication/authorization logic
- **Separation of Concerns**: Authentication (authn) and authorization (authz) are independent and can be used separately
- **Type Safety**: Context carries typed user information

## Directory Structure

```
internal/pkg/auth/
├── authenticator.go      # Authenticator core (token verification + UserLoader)
├── authenticator_test.go # Authenticator tests
├── authorizer.go         # Authorizer core (Casbin wrapper + SubjectResolver)
├── interceptor.go        # gRPC interceptors
├── middleware.go         # HTTP middleware
├── password.go           # Password encryption/comparison
└── websocket.go          # WebSocket helpers

internal/apiserver/biz/auth/
└── loader.go             # apiserver UserLoader implementation

internal/admserver/biz/auth/
└── loader.go             # admserver AdminLoader + SubjectResolver implementation
```

## Core Interfaces

### UserLoader - Load User Information

```go
// internal/pkg/auth/authenticator.go

// UserLoader loads user information into context
type UserLoader interface {
    LoadUser(ctx context.Context, userID string) (context.Context, error)
}

// Authenticator unified authenticator
type Authenticator struct {
    loader UserLoader
}

// New creates an authenticator, loader is optional
func New(loader UserLoader) *Authenticator {
    return &Authenticator{loader: loader}
}

// Verify validates token and loads user information
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

### SubjectResolver - Get Authorization Subject

```go
// internal/pkg/auth/authorizer.go

// SubjectResolver gets authorization subject from context
type SubjectResolver interface {
    ResolveSubject(ctx context.Context) (string, error)
}

// Authorizer unified authorizer
type Authorizer struct {
    enforcer *casbin.Enforcer
    resolver SubjectResolver
}

// NewAuthorizer creates an authorizer
func NewAuthorizer(db *gorm.DB, resolver SubjectResolver) (*Authorizer, error) {
    // Initialize casbin...
}

// Authorize checks permissions
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

## Service Implementation Examples

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

// AdminSubjectResolver gets authorization subject from admin info
type AdminSubjectResolver struct{}

func (r *AdminSubjectResolver) ResolveSubject(ctx context.Context) (string, error) {
    admin, ok := contextx.UserInfo[v1.AdminInfo](ctx)
    if !ok {
        return "", errorsx.New(401, "Unauthenticated", "admin info not found")
    }

    return known.RolePrefix + admin.RoleName, nil
}
```

## Usage

### apiserver HTTP (router/api.go)

```go
func MapApiRouters(g *gin.Engine) {
    loader := apiauth.NewUserLoader(store.S)
    authn := auth.New(loader)

    v1 := g.Group("/v1")

    // Public routes...

    v1.Use(auth.Middleware(authn))

    // Routes requiring authentication...
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

    // Public routes...

    v1.Use(auth.Middleware(authn))

    // Routes requiring authentication...

    v1.Use(auth.AuthzMiddleware(authz))

    // Routes requiring authorization...
}
```

## Related Documentation

- [Pluggable Protocol Layer](protocol-layer.md) - HTTP/gRPC/WebSocket unified architecture
- [Unified Error Handling](unified-error-handling.md) - Consistent error format across all protocols
