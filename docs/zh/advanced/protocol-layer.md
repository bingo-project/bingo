# 可插拔协议层

本文档介绍 Bingo 的可插拔协议层设计，支持 HTTP/gRPC/WebSocket 任意组合。

## 设计目标

1. **分层清晰** - Clean Architecture，协议层与业务层解耦
2. **可插拔** - 协议层独立可选，配置驱动
3. **统一格式** - 错误响应、认证机制三协议一致
4. **开发效率** - Biz 层写一次，多协议复用

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                     协议层 (可插拔)                           │
│  ┌────────┐  ┌────────┐  ┌─────────┐  ┌────────┐           │
│  │  HTTP  │  │  gRPC  │  │ Gateway │  │   WS   │           │
│  │Handler │  │Handler │  │ (可选)  │  │Handler │           │
│  └───┬────┘  └───┬────┘  └────┬────┘  └───┬────┘           │
└──────┼───────────┼────────────┼───────────┼────────────────┘
       └───────────┴─────┬──────┴───────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                       Biz 层                                 │
│                    (协议无关的业务逻辑)                        │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Store 层                                │
└─────────────────────────────────────────────────────────────┘
```

## 目录结构

```
internal/pkg/server/        # 通用服务基础设施（可复用）
├── server.go               # Server 接口, Runner
├── http.go                 # HTTP Server
├── grpc.go                 # gRPC Server
├── gateway.go              # gRPC-Gateway Server
├── websocket.go            # WebSocket Server
└── assembler.go            # 配置驱动组装

internal/apiserver/
├── biz/                    # 业务逻辑（协议无关）
│   ├── biz.go              # interface 定义
│   └── user/
│       └── user.go         # 实现（参数类型见「Biz 层参数类型」章节）
│
├── handler/                # 协议处理器（可插拔）
│   ├── http/               # 独立 HTTP（Gin）
│   ├── grpc/               # gRPC Handler
│   └── ws/                 # WebSocket (JSON-RPC 2.0)
│
└── store/                  # 数据访问

pkg/ws/                     # WebSocket 基础设施
├── hub.go                  # 连接管理
└── client.go               # 客户端连接

pkg/jsonrpc/                # JSON-RPC 2.0 支持
├── message.go              # 消息类型
├── response.go             # 响应构造
└── adapter.go              # 方法路由适配器

pkg/proto/                  # Proto 定义（数据层）
├── user/v1/
│   ├── user.proto          # 消息定义（所有模式共用）
│   └── user_service.proto  # 服务定义（gRPC/Gateway 用）
└── common/
    └── error.proto         # 统一错误类型（可选）
```

**说明**：`internal/pkg/server/` 是通用的服务基础设施，可被所有服务复用（apiserver、admserver 等）。

## 配置驱动

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

**mode 说明**：
- `standalone`：独立 HTTP，直接调用 Biz
- `gateway`：gRPC-Gateway 模式，代理到 gRPC（需要 grpc.enabled = true）

## 服务启动逻辑

### Server 接口

```go
// internal/pkg/server/server.go

// Server 可插拔服务器接口
type Server interface {
    Run(ctx context.Context) error      // 启动（阻塞直到 ctx 取消）
    Shutdown(ctx context.Context) error // 优雅关闭
    Name() string                       // 服务名称（用于日志）
}

// Runner 管理多个服务的生命周期
type Runner struct {
    servers []Server
}

func (r *Runner) Run(ctx context.Context) error    // 并发启动，任一失败触发全部关闭
func (r *Runner) Shutdown(ctx context.Context) error // 逆序关闭
```

### Assembler 组装

```go
// internal/pkg/server/assembler.go

// Option 模式配置
func WithGinEngine(engine *gin.Engine) AssemblerOption      // HTTP
func WithGRPCServer(server *grpc.Server) AssemblerOption    // gRPC
func WithWebSocket(engine *gin.Engine, hub *ws.Hub) AssemblerOption // WebSocket

// 根据配置组装
func Assemble(cfg *config.Config, opts ...AssemblerOption) *Runner
```

### 调用示例

```go
// internal/apiserver/apiserver.go

func Run(cfg *config.Config) error {
    // 1. 初始化依赖
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

    // 5. 组装并运行
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

**设计原则**：
- **依赖注入**：Handler 由调用方创建，Server 只管生命周期
- **单一职责**：`internal/pkg/server` 不依赖任何业务代码
- **可复用**：所有服务（apiserver、admserver 等）共用同一套基础设施

## Biz 层参数类型

Biz 层的方法参数和返回值支持两种方案：

### 方案一：Go Struct（本项目示例）

```go
// pkg/api/apiserver/v1/auth.go
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

// internal/apiserver/biz/auth/auth.go
func (b *authBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error)
```

**特点**：
- 使用 `binding` tag 做请求验证，Gin 原生支持
- 支持 Go 原生类型（`time.Time`、`*string` 可选字段等）
- gRPC Handler 需要做 Go Struct ↔ Proto Message 转换

### 方案二：Proto Message

```go
// 由 proto 文件生成
// internal/apiserver/biz/auth/auth.go
func (b *authBiz) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error)
```

**特点**：
- Proto 文件作为唯一类型定义
- gRPC Handler 无需类型转换
- HTTP Handler 需要 JSON ↔ Proto 转换
- 验证逻辑需使用 protoc-gen-validate 或手写

## Handler 实现

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

### HTTP Handler（独立模式）

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

WebSocket 采用 JSON-RPC 2.0 规范和中间件架构，详见 [WebSocket 设计与实现](websocket.md)。

```go
// internal/apiserver/router/ws.go

func RegisterWSHandlers(router *ws.Router) {
    h := wshandler.NewHandler(store.S)

    // 全局中间件
    router.Use(middleware.Recovery, middleware.RequestID, middleware.Logger)

    // 公开方法
    public := router.Group()
    public.Handle("auth.login", h.Login, middleware.LoginStateUpdater)

    // 需要认证的方法
    private := router.Group(middleware.Auth)
    private.Handle("auth.user-info", h.UserInfo)
}
```

## 支持的组合

| 组合 | 配置 |
|-----|------|
| 纯 HTTP | http.enabled=true |
| 纯 gRPC | grpc.enabled=true |
| 纯 WS | websocket.enabled=true |
| HTTP + gRPC | http.enabled=true, grpc.enabled=true |
| gRPC-Gateway | http.enabled=true, http.mode=gateway, grpc.enabled=true |
| HTTP + WS | http.enabled=true, websocket.enabled=true |
| 全部 | 全部 enabled |

## 新增 API 工作量

| 协议模式 | 新增一个 API 需要做什么 |
|---------|----------------------|
| 纯 gRPC | 改 proto → 实现 Biz 方法 → 完成 |
| 纯 HTTP | 改 proto → 实现 Biz 方法 → 加路由 → 完成 |
| gRPC-Gateway | 改 proto（加 http 注解）→ 实现 Biz 方法 → 完成 |
| HTTP + gRPC | 改 proto → 实现 Biz 方法 → 加 HTTP 路由 → 完成 |
| 含 WebSocket | 以上 + 注册一行 |

## Proto 生成策略

| 模式 | 生成内容 | Makefile 命令 |
|-----|---------|--------------|
| 纯 HTTP | `*.pb.go` | `make gen.proto.msg` |
| 纯 gRPC | `*.pb.go` + `*_grpc.pb.go` | `make gen.proto.grpc` |
| gRPC-Gateway | `*.pb.go` + `*_grpc.pb.go` + `*.pb.gw.go` | `make gen.proto.gateway` |

## 相关文档

- [WebSocket 设计与实现](websocket.md) - JSON-RPC 2.0 消息格式、中间件架构、连接管理
- [gRPC-Gateway 集成](grpc-gateway.md) - Gateway 模式配置与使用
- [统一错误处理](unified-error-handling.md) - 三协议错误格式统一
- [统一认证授权](unified-auth.md) - 插件式认证授权架构

---

**下一步**：了解 [WebSocket 设计与实现](websocket.md)，深入 JSON-RPC 2.0 协议和中间件架构。
