# Protocol Layer 实施计划

基于 [protocol-layer.md](../zh/advanced/protocol-layer.md) 设计，分阶段实现可插拔协议层。

## 现状分析

| 组件 | 状态 | 说明 |
|-----|------|------|
| pkg/errorsx | ✅ 已有 | 有 `GRPCStatus()`，需添加 `JSONRPCCode()` |
| pkg/jsonrpc | ❌ 缺失 | 需新建 |
| internal/apiserver/biz | ✅ 已有 | 业务逻辑层 |
| internal/apiserver/controller | ✅ 已有 | HTTP 控制器（保留，不迁移） |
| internal/apiserver/grpc | ✅ 已有 | gRPC Handler（保留，不迁移） |
| internal/apiserver/handler/ws | ❌ 缺失 | 需新建 |
| internal/apiserver/handler/gateway | ❌ 缺失 | 需新建 |
| pkg/ws (feature/ws 分支) | ⚠️ 部分可用 | 复用基础设施，重构协议层 |
| 配置驱动服务组装 | ❌ 缺失 | 需重构 server.go |

## WebSocket 代码复用策略

从 `feature/ws` 分支复用以下代码：

| 文件 | 处理方式 | 说明 |
|-----|---------|------|
| `pkg/ws/cache/*.go` | ✅ 直接复用 | 用户在线状态缓存 |
| `pkg/ws/task/*.go` | ✅ 直接复用 | 定时任务（清理连接、服务注册） |
| `pkg/ws/server/client_hub.go` | ✅ 直接复用 | 连接管理器 |
| `pkg/ws/server/client.go` | ⚠️ 重构 | 保留 readPump/writePump，改消息处理 |
| `pkg/ws/server/register.go` | ❌ 删除 | 用 `pkg/jsonrpc.Adapter` 替代 |
| `pkg/ws/common/*.go` | ❌ 删除 | 用 `pkg/errorsx` 替代 |
| `pkg/ws/model/*.go` | ❌ 删除 | 用 `pkg/jsonrpc` + proto 替代 |
| `pkg/ws/helper/*.go` | ⚠️ 评估 | 按需保留 |

## 阶段一：JSON-RPC 基础设施

### 1.1 添加 JSONRPCCode 方法

**文件**: `pkg/errorsx/jsonrpc.go`

```go
// ABOUTME: HTTP to JSON-RPC error code mapping.
// ABOUTME: Provides JSONRPCCode() method for ErrorX.

package errorsx

// HTTP → JSON-RPC 错误码映射
var httpToJSONRPC = map[int]int{
    400: -32602, // Invalid params
    401: -32001, // Unauthenticated
    403: -32003, // Permission denied
    404: -32004, // Not found
    409: -32009, // Conflict
    429: -32029, // Too many requests
    500: -32603, // Internal error
    503: -32053, // Service unavailable
}

// JSONRPCCode 返回 JSON-RPC 错误码
func (err *ErrorX) JSONRPCCode() int {
    if code, ok := httpToJSONRPC[err.Code]; ok {
        return code
    }
    return -32603 // 默认 Internal error
}
```

### 1.2 创建 pkg/jsonrpc 包

**文件**: `pkg/jsonrpc/message.go`

```go
// ABOUTME: JSON-RPC 2.0 message types for WebSocket communication.
// ABOUTME: Defines Request, Response, and Error structures per JSON-RPC 2.0 spec.

package jsonrpc

import "encoding/json"

const Version = "2.0"

// Request JSON-RPC 2.0 请求
type Request struct {
    JSONRPC string          `json:"jsonrpc"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
    ID      any             `json:"id,omitempty"`
}

// Response JSON-RPC 2.0 响应
type Response struct {
    JSONRPC string `json:"jsonrpc"`
    Result  any    `json:"result,omitempty"`
    Error   *Error `json:"error,omitempty"`
    ID      any    `json:"id"`
}

// Error JSON-RPC 2.0 错误
type Error struct {
    Code    int               `json:"code"`
    Reason  string            `json:"reason"`
    Message string            `json:"message"`
    Data    map[string]string `json:"data,omitempty"`
}
```

**文件**: `pkg/jsonrpc/response.go`

```go
// ABOUTME: JSON-RPC 2.0 response constructors.
// ABOUTME: Creates success responses, error responses, and notifications.

package jsonrpc

import "bingo/pkg/errorsx"

// NewResponse 创建成功响应
func NewResponse(id any, result any) *Response {
    return &Response{
        JSONRPC: Version,
        Result:  result,
        ID:      id,
    }
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(id any, err error) *Response {
    e := errorsx.FromError(err)
    return &Response{
        JSONRPC: Version,
        Error: &Error{
            Code:    e.JSONRPCCode(),
            Reason:  e.Reason,
            Message: e.Message,
            Data:    e.Metadata,
        },
        ID: id,
    }
}

// NewStreamResponse 创建流式响应（有 id，带 method 标识）
func NewStreamResponse(id any, method string, result any) *Response {
    return &Response{
        JSONRPC: Version,
        Method:  method,
        Result:  result,
        ID:      id,
    }
}

// NewPush 创建服务端主动推送（无 id，不关联请求）
func NewPush(method string, data any) *Push {
    return &Push{
        JSONRPC: Version,
        Method:  method,
        Data:    data,
    }
}
```

## 阶段二：JSON-RPC 适配器

### 2.1 创建适配器

**文件**: `pkg/jsonrpc/adapter.go`

```go
// ABOUTME: JSON-RPC to Biz layer adapter.
// ABOUTME: Routes JSON-RPC methods to proto-based handlers.

package jsonrpc

import (
    "context"
    "encoding/json"
    "reflect"

    "google.golang.org/protobuf/proto"

    "bingo/pkg/errorsx"
)

// HandlerFunc Handler 函数签名
type HandlerFunc func(ctx context.Context, req proto.Message) (proto.Message, error)

// Adapter JSON-RPC 到 Biz 层的适配器
type Adapter struct {
    handlers map[string]*handlerInfo
}

type handlerInfo struct {
    handler     HandlerFunc
    requestType reflect.Type
}

func NewAdapter() *Adapter {
    return &Adapter{
        handlers: make(map[string]*handlerInfo),
    }
}

// Register 注册 Handler
func (a *Adapter) Register(method string, handler HandlerFunc, reqType proto.Message) {
    a.handlers[method] = &handlerInfo{
        handler:     handler,
        requestType: reflect.TypeOf(reqType).Elem(),
    }
}

// Handle 处理 JSON-RPC 请求
func (a *Adapter) Handle(ctx context.Context, req *Request) *Response {
    info, ok := a.handlers[req.Method]
    if !ok {
        return NewErrorResponse(req.ID,
            errorsx.New(404, "MethodNotFound", "Method not found: %s", req.Method))
    }

    protoReq := reflect.New(info.requestType).Interface().(proto.Message)

    if len(req.Params) > 0 {
        if err := json.Unmarshal(req.Params, protoReq); err != nil {
            return NewErrorResponse(req.ID,
                errorsx.New(400, "InvalidParams", "Invalid params: %s", err.Error()))
        }
    }

    resp, err := info.handler(ctx, protoReq)
    if err != nil {
        return NewErrorResponse(req.ID, err)
    }

    data, _ := json.Marshal(resp)
    var result any
    json.Unmarshal(data, &result)

    return NewResponse(req.ID, result)
}
```

### 2.2 泛型注册

**文件**: `pkg/jsonrpc/generic.go`

```go
// ABOUTME: Generic handler registration for JSON-RPC adapter.
// ABOUTME: Provides type-safe registration using Go generics.

package jsonrpc

import (
    "context"

    "google.golang.org/protobuf/proto"
)

// Register 泛型注册（Go 1.18+）
func Register[Req, Resp proto.Message](
    a *Adapter,
    method string,
    handler func(context.Context, Req) (Resp, error),
    reqType Req,
) {
    a.Register(method, func(ctx context.Context, req proto.Message) (proto.Message, error) {
        return handler(ctx, req.(Req))
    }, reqType)
}
```

## 阶段三：WebSocket Handler（新建）

### 3.1 复用基础设施

从 `feature/ws` 分支拷贝以下文件到 `pkg/ws/`：

```bash
# 直接复用
git checkout feature/ws -- pkg/ws/cache/
git checkout feature/ws -- pkg/ws/task/
git checkout feature/ws -- pkg/ws/server/client_hub.go

# 重构后使用
git show feature/ws:pkg/ws/server/client.go > pkg/ws/server/client.go.bak
```

### 3.2 统一 Handler 目录结构

迁移后的目录结构：

```
internal/apiserver/handler/
├── http/           # HTTP 控制器（从 controller/v1/ 迁移）
│   ├── auth/
│   ├── common/
│   └── file/
├── grpc/           # gRPC Handler（从 grpc/v1/apiserver/ 迁移）
└── ws/             # WebSocket Handler
    ├── handler.go
    └── router.go
```

**迁移步骤**：
1. 创建 `handler/http/`，移动 `controller/v1/` 内容（去掉 v1 层级）
2. 创建 `handler/grpc/`，移动 `grpc/v1/apiserver/` 内容（去掉 v1 层级）
3. 更新所有 import 路径
4. 删除旧目录 `controller/`、`grpc/`

### 3.3 重构 Client

**文件**: `pkg/ws/server/client.go`

```go
// ABOUTME: WebSocket client connection management.
// ABOUTME: Handles message read/write with JSON-RPC 2.0 protocol.

package server

import (
    "context"
    "encoding/json"
    "sync"
    "time"

    "github.com/gorilla/websocket"

    "bingo/pkg/jsonrpc"
    "bingo/pkg/log"
)

type Client struct {
    conn    *websocket.Conn
    send    chan []byte
    ctx     context.Context      // 携带认证信息
    adapter *jsonrpc.Adapter     // 消息处理适配器

    Addr          string
    AppID         uint32
    UserID        string
    FirstTime     uint64
    HeartbeatTime uint64
    LoginTime     uint64

    mu sync.RWMutex
}

func NewClient(conn *websocket.Conn, ctx context.Context, adapter *jsonrpc.Adapter) *Client {
    now := uint64(time.Now().Unix())
    return &Client{
        conn:          conn,
        send:          make(chan []byte, 100),
        ctx:           ctx,
        adapter:       adapter,
        Addr:          conn.RemoteAddr().String(),
        FirstTime:     now,
        HeartbeatTime: now,
    }
}

func (c *Client) handleMessage(data []byte) {
    var req jsonrpc.Request
    if err := json.Unmarshal(data, &req); err != nil {
        resp := jsonrpc.NewErrorResponse(nil,
            errorsx.New(400, "ParseError", "Invalid JSON"))
        c.SendJSON(resp)
        return
    }

    // 心跳特殊处理
    if req.Method == "heartbeat" {
        c.Heartbeat(uint64(time.Now().Unix()))
        c.SendJSON(jsonrpc.NewResponse(req.ID, nil))
        return
    }

    // 通过适配器调用 Biz 层
    resp := c.adapter.Handle(c.ctx, &req)
    c.SendJSON(resp)
}

func (c *Client) SendJSON(v any) {
    data, err := json.Marshal(v)
    if err != nil {
        log.Errorw("JSON marshal error", "err", err)
        return
    }
    c.send <- data
}

// readPump/writePump 保持原有实现...
```

### 3.4 WebSocket Handler

**文件**: `internal/apiserver/handler/ws/handler.go`

```go
// ABOUTME: WebSocket HTTP handler for Gin.
// ABOUTME: Upgrades HTTP connections and manages WebSocket lifecycle.

package ws

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"

    "bingo/internal/pkg/contextx"
    "bingo/pkg/jsonrpc"
    "bingo/pkg/token"
    "bingo/pkg/ws/server"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin:     func(r *http.Request) bool { return true },
}

type Handler struct {
    adapter *jsonrpc.Adapter
}

func NewHandler(adapter *jsonrpc.Adapter) *Handler {
    return &Handler{adapter: adapter}
}

func (h *Handler) ServeWs(c *gin.Context) {
    // 1. 获取 Token
    tokenStr := c.Query("token")
    if tokenStr == "" {
        tokenStr = extractBearerToken(c.GetHeader("Authorization"))
    }

    // 2. 创建 context
    ctx := context.Background()
    ctx = contextx.WithRequestID(ctx, c.GetHeader("X-Request-ID"))
    ctx = contextx.WithClientIP(ctx, c.ClientIP())

    // 3. 认证（如果有 token）
    if tokenStr != "" {
        claims, err := token.Parse(tokenStr)
        if err == nil {
            ctx = contextx.WithUserID(ctx, claims.Subject)
        }
    }

    // 4. 升级连接
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }

    // 5. 创建客户端
    client := server.NewClient(conn, ctx, h.adapter)

    // 6. 注册到 ClientManager
    if userID := contextx.UserID(ctx); userID != "" {
        client.SetUserInfo(userID, 0)
        server.ClientManager.Login <- client
    } else {
        server.ClientManager.Register <- client
    }

    // 7. 启动读写协程
    go client.WritePump()
    go client.ReadPump()
}

func extractBearerToken(auth string) string {
    const prefix = "Bearer "
    if len(auth) > len(prefix) && auth[:len(prefix)] == prefix {
        return auth[len(prefix):]
    }
    return ""
}
```

### 3.5 方法注册

**文件**: `internal/apiserver/handler/ws/router.go`

```go
// ABOUTME: WebSocket method registration for JSON-RPC.
// ABOUTME: Maps JSON-RPC methods to Biz layer handlers.

package ws

import (
    "bingo/internal/apiserver/biz"
    "bingo/pkg/jsonrpc"
    pb "bingo/pkg/proto/apiserver/v1"
)

func RegisterHandlers(a *jsonrpc.Adapter, b biz.IBiz) {
    jsonrpc.Register(a, "user.login", b.Auth().Login, &pb.LoginRequest{})
    jsonrpc.Register(a, "user.logout", b.Auth().Logout, &pb.LogoutRequest{})
    // ... 其他方法
}
```

## 阶段四：配置驱动

### 4.1 Server 接口定义

**文件**: `internal/apiserver/server/server.go`

```go
// ABOUTME: Pluggable server interface and runner.
// ABOUTME: Enables configuration-driven protocol selection with graceful shutdown.

package server

import (
    "context"
    "errors"
    "fmt"
    "time"

    "golang.org/x/sync/errgroup"
)

// Server 可插拔服务器接口
type Server interface {
    // Run 启动服务器（阻塞直到 context 取消或错误）
    Run(ctx context.Context) error
    // Shutdown 优雅关闭
    Shutdown(ctx context.Context) error
    // Name 返回服务器名称（用于日志）
    Name() string
}

// Runner 服务运行器，管理多个服务的生命周期
type Runner struct {
    servers []Server
}

// NewRunner 创建服务运行器
func NewRunner(servers ...Server) *Runner {
    return &Runner{servers: servers}
}

// Run 启动所有服务，任一失败则触发全部关闭
func (r *Runner) Run(ctx context.Context) error {
    if len(r.servers) == 0 {
        return nil
    }

    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    g, ctx := errgroup.WithContext(ctx)

    // 启动所有服务
    for _, srv := range r.servers {
        srv := srv
        g.Go(func() error {
            return srv.Run(ctx)
        })
    }

    // 监听 context 取消，触发优雅关闭
    g.Go(func() error {
        <-ctx.Done()
        // 使用独立 context 进行关闭，避免立即超时
        shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer shutdownCancel()
        return r.Shutdown(shutdownCtx)
    })

    return g.Wait()
}

// Shutdown 优雅关闭（逆序关闭，收集所有错误）
func (r *Runner) Shutdown(ctx context.Context) error {
    var errs []error

    // 逆序关闭：后启动的先关闭
    for i := len(r.servers) - 1; i >= 0; i-- {
        srv := r.servers[i]
        if err := srv.Shutdown(ctx); err != nil {
            errs = append(errs, fmt.Errorf("%s: %w", srv.Name(), err))
        }
    }

    return errors.Join(errs...)
}
```

**关闭顺序说明**：
- 服务按添加顺序启动：HTTP → gRPC → WebSocket
- 服务按逆序关闭：WebSocket → gRPC → HTTP
- 这确保长连接（WebSocket）先关闭，然后是 RPC 服务，最后是入口服务

### 4.2 配置结构

**文件**: `internal/apiserver/config/server.go`

```go
// ABOUTME: Server configuration for protocol layer.
// ABOUTME: Supports HTTP, gRPC, WebSocket, and Gateway modes.

package config

type ServerConfig struct {
    HTTP      HTTPConfig      `json:"http" mapstructure:"http"`
    GRPC      GRPCConfig      `json:"grpc" mapstructure:"grpc"`
    WebSocket WebSocketConfig `json:"websocket" mapstructure:"websocket"`
}

type HTTPConfig struct {
    Enabled bool   `json:"enabled" mapstructure:"enabled"`
    Addr    string `json:"addr" mapstructure:"addr"`
    Mode    string `json:"mode" mapstructure:"mode"` // standalone | gateway
}

type GRPCConfig struct {
    Enabled bool   `json:"enabled" mapstructure:"enabled"`
    Addr    string `json:"addr" mapstructure:"addr"`
}

type WebSocketConfig struct {
    Enabled bool   `json:"enabled" mapstructure:"enabled"`
    Addr    string `json:"addr" mapstructure:"addr"`
}
```

### 4.3 服务组装

**文件**: `internal/apiserver/server/assembler.go`

```go
// ABOUTME: Server assembler based on configuration.
// ABOUTME: Creates Runner with enabled servers according to config.

package server

import (
    "bingo/internal/apiserver/biz"
    "bingo/internal/apiserver/config"
    "bingo/internal/apiserver/handler/gateway"
    "bingo/internal/apiserver/handler/ws"
    grpchandler "bingo/internal/apiserver/handler/grpc"
    httphandler "bingo/internal/apiserver/handler/http"
)

// Assemble 根据配置组装服务运行器
func Assemble(cfg *config.Config, biz biz.IBiz) *Runner {
    var servers []Server

    // 启动顺序：HTTP/Gateway → gRPC → WebSocket
    // 关闭顺序（逆序）：WebSocket → gRPC → HTTP/Gateway

    // 1. HTTP 或 Gateway
    if cfg.Server.HTTP.Enabled {
        switch cfg.Server.HTTP.Mode {
        case "gateway":
            if cfg.Server.GRPC.Enabled {
                servers = append(servers, gateway.NewServer(cfg))
            }
        default:
            servers = append(servers, httphandler.NewServer(cfg, biz))
        }
    }

    // 2. gRPC
    if cfg.Server.GRPC.Enabled {
        servers = append(servers, grpchandler.NewServer(cfg, biz))
    }

    // 3. WebSocket
    if cfg.Server.WebSocket.Enabled {
        servers = append(servers, ws.NewServer(cfg, biz))
    }

    return NewRunner(servers...)
}
```

**使用示例**：

```go
// internal/apiserver/run.go

func run(cfg *config.Config) error {
    // 初始化依赖
    storeInstance := store.NewStore(db)
    bizInstance := biz.NewBiz(storeInstance)

    // 组装并运行
    runner := server.Assemble(cfg, bizInstance)

    // 监听系统信号
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    return runner.Run(ctx)
}
```

## 阶段五：gRPC-Gateway

### 5.1 Proto 添加 HTTP 注解

**文件**: `pkg/proto/apiserver/v1/auth.proto`

```protobuf
import "google/api/annotations.proto";

service AuthService {
    rpc Login(LoginRequest) returns (LoginResponse) {
        option (google.api.http) = {
            post: "/v1/auth/login"
            body: "*"
        };
    }
}
```

### 5.2 生成 Gateway 代码

**文件**: `Makefile`

```makefile
.PHONY: gen.proto.gateway
gen.proto.gateway:
	protoc --go_out=. --go-grpc_out=. \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=logtostderr=true \
		pkg/proto/apiserver/v1/*.proto
```

### 5.3 Gateway Server

**文件**: `internal/apiserver/handler/gateway/server.go`

```go
// ABOUTME: gRPC-Gateway HTTP server.
// ABOUTME: Proxies HTTP requests to gRPC backend.

package gateway

import (
    "context"
    "net/http"

    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    "bingo/internal/apiserver/config"
    pb "bingo/pkg/proto/apiserver/v1"
)

type Server struct {
    cfg    *config.Config
    server *http.Server
}

func NewServer(cfg *config.Config) *Server {
    return &Server{cfg: cfg}
}

func (s *Server) Name() string {
    return "grpc-gateway"
}

func (s *Server) Run() error {
    ctx := context.Background()
    mux := runtime.NewServeMux()

    opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

    // 注册所有服务
    if err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, s.cfg.Server.GRPC.Addr, opts); err != nil {
        return err
    }

    s.server = &http.Server{
        Addr:    s.cfg.Server.HTTP.Addr,
        Handler: mux,
    }

    return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    return s.server.Shutdown(ctx)
}
```

## 阶段六：统一认证

### 6.1 认证器接口

**文件**: `internal/pkg/auth/authenticator.go`

```go
// ABOUTME: Unified authenticator for all protocols.
// ABOUTME: Provides middleware/interceptor for HTTP, gRPC, and WebSocket.

package auth

import (
    "context"

    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"

    "bingo/internal/pkg/contextx"
    "bingo/pkg/token"
)

type Authenticator struct {
    // 可扩展：添加 UserGetter 等依赖
}

func New() *Authenticator {
    return &Authenticator{}
}

// Verify 验证 token 并返回带用户信息的 context
func (a *Authenticator) Verify(ctx context.Context, tokenStr string) (context.Context, error) {
    claims, err := token.Parse(tokenStr)
    if err != nil {
        return ctx, err
    }

    ctx = contextx.WithUserID(ctx, claims.Subject)
    return ctx, nil
}

// HTTPMiddleware Gin 中间件
func (a *Authenticator) HTTPMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenStr := extractBearerToken(c.GetHeader("Authorization"))
        if tokenStr == "" {
            c.Next()
            return
        }

        ctx, err := a.Verify(c.Request.Context(), tokenStr)
        if err != nil {
            c.Next()
            return
        }

        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}

// GRPCInterceptor gRPC 拦截器
func (a *Authenticator) GRPCInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
        // 从 metadata 获取 token
        // ...
        return handler(ctx, req)
    }
}

// WSVerify WebSocket 认证
func (a *Authenticator) WSVerify(ctx context.Context, tokenStr string) (context.Context, error) {
    return a.Verify(ctx, tokenStr)
}
```

## 实施顺序

| 阶段 | 任务 | 依赖 | 预估复杂度 |
|-----|------|------|-----------|
| 1.1 | 添加 JSONRPCCode 方法 | 无 | 低 |
| 1.2 | 创建 pkg/jsonrpc 包 | 1.1 | 低 |
| 2.1 | 创建 JSON-RPC 适配器 | 1.2 | 中 |
| 2.2 | 添加泛型注册 | 2.1 | 低 |
| 3.1 | 复用 ws 基础设施 | 无 | 低 |
| 3.2 | 重构 Client | 2.2, 3.1 | 中 |
| 3.3 | 实现 WebSocket Handler | 3.2 | 中 |
| 3.4 | 注册业务方法 | 3.3 | 低 |
| 4.1 | 定义 Server 接口 | 无 | 低 |
| 4.2 | 扩展配置结构 | 无 | 低 |
| 4.3 | 实现服务组装 | 3.3, 4.1, 4.2 | 中 |
| 5.1 | Proto 添加 HTTP 注解 | 无 | 低 |
| 5.2 | 生成 Gateway 代码 | 5.1 | 低 |
| 5.3 | 实现 Gateway Server | 5.2, 4.1 | 中 |
| 6.1 | 统一认证 | 4.3 | 中 |

## 验收标准

- [x] `pkg/errorsx.JSONRPCCode()` 通过单元测试
- [x] `pkg/jsonrpc` 消息编解码通过单元测试
- [x] WebSocket 连接可以成功调用业务方法
- [x] 配置 `websocket.enabled=false` 时不启动 WebSocket 服务
- [x] 配置 `http.mode=gateway` 时使用 gRPC-Gateway
- [x] 三种协议使用相同的认证逻辑
- [x] 三种协议返回相同格式的错误响应
- [x] Handler 目录结构统一（controller/, grpc/ 迁移到 handler/）
- [ ] Server 包提取到 internal/pkg/server（支持多服务复用）

## 实施记录

| 阶段 | 完成日期 | Commit |
|-----|---------|--------|
| Phase 1: JSON-RPC 基础设施 | 2024-12 | `0d56eaf` |
| Phase 2: JSON-RPC 适配器 | 2024-12 | `0d56eaf` |
| Phase 3: WebSocket Handler | 2024-12 | `408e693` |
| Phase 4: Config-driven 服务组装 | 2024-12 | `6de4b10` |
| Phase 5: gRPC-Gateway | 2024-12 | `dba9b6b` |
| Phase 6: 统一认证 | 2024-12 | `7b1da9f` |
| Code Review 修复 | 2024-12 | `2d1a8ef` |
| 配置增强 (Origin/TLS) | 2024-12 | `58b1078` |

## Code Review

详细报告见 [protocol-layer-review.md](../reviews/protocol-layer-review.md)

### 已修复问题

| 问题 | 严重性 | 修复内容 |
|------|--------|----------|
| Gateway 使用不安全凭证 | 关键 | 添加 TLS 配置支持 |
| Hub.Run() 无法优雅停止 | 关键 | 添加 context 支持 |
| Send channel 资源泄漏 | 重要 | 在 unregister 时关闭 channel |
| JSON marshal 错误被忽略 | 重要 | 正确处理并返回错误响应 |
| handleMessage 无 panic recovery | 重要 | 添加 defer recover |
| WebSocket Origin 验证 | 重要 | 添加 allowedOrigins 配置 |
| gRPC TLS 支持 | 重要 | 添加 TLS 配置 (secure/insecure) |

### 待改进项（非阻塞）

- [ ] 类型系统迁移到 proto.Message（已有 TODO）
