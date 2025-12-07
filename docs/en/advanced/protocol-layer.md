# Pluggable Protocol Layer

This document describes Bingo's pluggable protocol layer design, supporting any combination of HTTP/gRPC/WebSocket.

## Design Goals

1. **Clean Architecture** - Protocol layer decoupled from business layer
2. **Pluggable** - Protocols are independent and configurable
3. **Unified Format** - Consistent error responses and authentication across all protocols
4. **Development Efficiency** - Write Biz layer once, reuse across multiple protocols

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Protocol Layer (Pluggable)               │
│  ┌────────┐  ┌────────┐  ┌─────────┐  ┌────────┐           │
│  │  HTTP  │  │  gRPC  │  │ Gateway │  │   WS   │           │
│  │Handler │  │Handler │  │(Optional)│  │Handler │           │
│  └───┬────┘  └───┬────┘  └────┬────┘  └───┬────┘           │
└──────┼───────────┼────────────┼───────────┼────────────────┘
       └───────────┴─────┬──────┴───────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                        Biz Layer                            │
│              (Proto message as params/returns)              │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                       Store Layer                           │
└─────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
internal/pkg/server/        # Common server infrastructure (reusable)
├── server.go               # Server interface, Runner
├── http.go                 # HTTP Server
├── grpc.go                 # gRPC Server
├── gateway.go              # gRPC-Gateway Server
├── websocket.go            # WebSocket Server
└── assembler.go            # Configuration-driven assembly

internal/apiserver/
├── biz/                    # Business logic (protocol-agnostic)
│   ├── biz.go              # Interface definitions
│   └── user/
│       └── user.go         # Implementation, uses proto messages
│
├── handler/                # Protocol handlers (pluggable)
│   ├── http/               # Standalone HTTP (Gin)
│   ├── grpc/               # gRPC Handler
│   └── ws/                 # WebSocket (JSON-RPC 2.0)
│
└── store/                  # Data access

pkg/ws/                     # WebSocket infrastructure
├── hub.go                  # Connection management
└── client.go               # Client connection

pkg/jsonrpc/                # JSON-RPC 2.0 support
├── message.go              # Message types
├── response.go             # Response construction
└── adapter.go              # Method routing adapter

pkg/proto/                  # Proto definitions (data layer)
├── user/v1/
│   ├── user.proto          # Message definitions (shared by all modes)
│   └── user_service.proto  # Service definitions (for gRPC/Gateway)
└── common/
    └── error.proto         # Unified error type (optional)
```

**Note**: `internal/pkg/server/` is common server infrastructure, reusable by all services (apiserver, admserver, etc.).

## Configuration-Driven

```yaml
# configs/apiserver.yaml
server:
  http:
    enabled: true
    addr: ":8080"
    mode: standalone    # standalone | gateway
  grpc:
    enabled: true
    addr: ":9090"
  websocket:
    enabled: true
    addr: ":8081"
```

**Mode Options**:
- `standalone`: Standalone HTTP, calls Biz directly
- `gateway`: gRPC-Gateway mode, proxies to gRPC (requires grpc.enabled = true)

## Server Startup Logic

### Server Interface

```go
// internal/pkg/server/server.go

// Server pluggable server interface
type Server interface {
    Run(ctx context.Context) error      // Start (blocks until ctx cancelled)
    Shutdown(ctx context.Context) error // Graceful shutdown
    Name() string                       // Server name (for logging)
}

// Runner manages lifecycle of multiple servers
type Runner struct {
    servers []Server
}

func (r *Runner) Run(ctx context.Context) error    // Start concurrently, any failure triggers shutdown
func (r *Runner) Shutdown(ctx context.Context) error // Shutdown in reverse order
```

### Assembler

```go
// internal/pkg/server/assembler.go

// Option pattern configuration
func WithGinEngine(engine *gin.Engine) AssemblerOption      // HTTP
func WithGRPCServer(server *grpc.Server) AssemblerOption    // gRPC
func WithWebSocket(engine *gin.Engine, hub *ws.Hub) AssemblerOption // WebSocket

// Assemble based on configuration
func Assemble(cfg *config.Config, opts ...AssemblerOption) *Runner
```

### Usage Example

```go
// internal/apiserver/apiserver.go

func Run(cfg *config.Config) error {
    // 1. Initialize dependencies
    bizInstance := biz.NewBiz(store.S)

    // 2. HTTP
    httpEngine := gin.New()
    router.MapHTTPRouters(httpEngine, bizInstance)

    // 3. gRPC
    grpcSrv := grpc.NewServer()
    router.MapGRPCRouters(grpcSrv, bizInstance)

    // 4. WebSocket
    hub := ws.NewHub()
    wsRouter := ws.NewRouter()
    router.RegisterWSHandlers(wsRouter)
    wsEngine := gin.New()
    wsEngine.GET("/ws", wshandler.ServeWS(hub, wsRouter))

    // 5. Assemble and run
    runner := server.Assemble(cfg,
        server.WithGinEngine(httpEngine),
        server.WithGRPCServer(grpcSrv),
        server.WithWebSocket(wsEngine, hub),
    )

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    return runner.Run(ctx)
}
```

**Design Principles**:
- **Dependency Injection**: Handlers created by caller, Server only manages lifecycle
- **Single Responsibility**: `internal/pkg/server` has no business code dependencies
- **Reusable**: All services (apiserver, admserver, etc.) share the same infrastructure

## Handler Implementation

### gRPC Handler

```go
type UserHandler struct {
    pb.UnimplementedUserServiceServer
    biz biz.IBiz
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
    return h.biz.User().Login(ctx, req)
}
```

### HTTP Handler (Standalone Mode)

```go
type UserHandler struct {
    biz biz.IBiz
}

func (h *UserHandler) Login(c *gin.Context) {
    var req pb.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
        return
    }

    resp, err := h.biz.User().Login(c.Request.Context(), &req)
    core.Response(c, resp, err)
}
```

### WebSocket Handler

WebSocket uses JSON-RPC 2.0 specification and middleware architecture. See [WebSocket Design and Implementation](websocket.md).

```go
// internal/apiserver/router/ws.go

func RegisterWSHandlers(router *ws.Router) {
    h := wshandler.NewHandler(store.S)

    // Global middleware
    router.Use(middleware.Recovery, middleware.RequestID, middleware.Logger)

    // Public methods
    public := router.Group()
    public.Handle("auth.login", h.Login, middleware.LoginStateUpdater)

    // Authenticated methods
    private := router.Group(middleware.Auth)
    private.Handle("auth.user-info", h.UserInfo)
}
```

## Supported Combinations

| Combination | Configuration |
|-------------|---------------|
| HTTP only | http.enabled=true |
| gRPC only | grpc.enabled=true |
| WS only | websocket.enabled=true |
| HTTP + gRPC | http.enabled=true, grpc.enabled=true |
| gRPC-Gateway | http.enabled=true, http.mode=gateway, grpc.enabled=true |
| HTTP + WS | http.enabled=true, websocket.enabled=true |
| All | All enabled |

## Adding New API

| Protocol Mode | Steps to Add an API |
|---------------|---------------------|
| gRPC only | Modify proto -> Implement Biz method -> Done |
| HTTP only | Modify proto -> Implement Biz method -> Add route -> Done |
| gRPC-Gateway | Modify proto (add http annotations) -> Implement Biz method -> Done |
| HTTP + gRPC | Modify proto -> Implement Biz method -> Add HTTP route -> Done |
| With WebSocket | Above + Register one line |

## Proto Generation Strategy

| Mode | Generated Content | Makefile Command |
|------|-------------------|------------------|
| HTTP only | `*.pb.go` | `make gen.proto.msg` |
| gRPC only | `*.pb.go` + `*_grpc.pb.go` | `make gen.proto.grpc` |
| gRPC-Gateway | `*.pb.go` + `*_grpc.pb.go` + `*.pb.gw.go` | `make gen.proto.gateway` |

## Related Documentation

- [WebSocket Design and Implementation](websocket.md) - JSON-RPC 2.0 message format, middleware architecture, connection management
- [gRPC-Gateway Integration](grpc-gateway.md) - Gateway mode configuration and usage
- [Unified Error Handling](unified-error-handling.md) - Consistent error format across all protocols
- [Unified Authentication](unified-auth.md) - Plugin-based authentication architecture

---

**Next Step**: Learn about [WebSocket Design and Implementation](websocket.md) for details on JSON-RPC 2.0 protocol and middleware architecture.
