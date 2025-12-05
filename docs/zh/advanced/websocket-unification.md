# WebSocket 与 HTTP/gRPC 统一方案

本文档介绍如何将 WebSocket 与 HTTP/gRPC 统一，实现三种协议共享同一套业务逻辑、错误处理和响应格式。

## 行业现状

经过调研，**目前没有一个完美的标准方案能让 HTTP/gRPC/WebSocket 三者完全统一**：

| 方案 | HTTP | gRPC | WebSocket 长连接 | 生产就绪 |
|-----|------|------|-----------------|---------|
| [ConnectRPC](https://connectrpc.com/) | ✅ | ✅ | ❌ (浏览器不支持 bidi) | ✅ |
| [gRPC-Gateway](https://github.com/grpc-ecosystem/grpc-gateway) | ✅ | ✅ | ❌ | ✅ |
| [grpc-websocket-proxy](https://github.com/tmc/grpc-websocket-proxy) | ✅ | ✅ | ⚠️ 实验性 | ⚠️ |

**核心问题：**
- gRPC 双向流依赖 HTTP/2
- 浏览器不支持 HTTP/2 bidi streaming
- WebSocket 是 HTTP/1.1 协议，与 gRPC 无法直接互通

**行业最佳实践：不强求协议层完全统一，而是统一数据结构、错误格式和业务逻辑。**

## 设计目标

1. **业务逻辑复用** - WebSocket 调用同一个 Biz 层
2. **数据结构统一** - 请求/响应使用相同的 Proto 或 Go struct
3. **错误格式统一** - 三种协议返回相同的错误结构
4. **认证机制统一** - WebSocket 连接认证复用相同的 token 验证逻辑

## 架构概览

### 推荐方案：WS → Biz 层

```
┌─────────────────────────────────────────────────────────────┐
│                 Proto / Go Struct (数据定义)                  │
│           LoginRequest / LoginResponse / ErrorX             │
└─────────────────────────────┬───────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
     ┌────────────┐    ┌────────────┐    ┌────────────┐
     │   HTTP     │    │   gRPC     │    │ WebSocket  │
     │ (Gateway)  │    │  Handler   │    │  Handler   │
     └─────┬──────┘    └─────┬──────┘    └─────┬──────┘
           │                 │                 │
           └─────────────────┼─────────────────┘
                             ▼
                    ┌────────────────┐
                    │    Biz 层      │  ← 三种协议共享
                    └────────────────┘
```

**统一的部分：**
- 数据结构（Proto 或 Go struct）
- 错误格式（`{"code": -32001, "message": "xxx", "data": {"reason": "xxx"}}`）
- 业务逻辑（Biz 层）

**各自独立的部分：**
- 协议适配层（各自的 Handler）
- WebSocket 使用 JSON-RPC 2.0 `method` 字段路由（HTTP 用 URL，gRPC 用方法名）

## 第一阶段：统一消息格式 (JSON-RPC 2.0)

### 为什么选择 JSON-RPC 2.0？

WebSocket 是传输协议，没有内置请求-响应匹配机制。[JSON-RPC 2.0](https://www.jsonrpc.org/specification) 是行业标准的 RPC 消息格式：

- **行业广泛采用**：[MCP](https://modelcontextprotocol.info/specification/2024-11-05/basic/messages/)、以太坊、VSCode LSP 等
- **传输无关**：可用于 WebSocket、HTTP、TCP 等任何传输层
- **客户端库丰富**：大量现成实现可用

### 1.1 JSON-RPC 2.0 请求格式

```go
// pkg/jsonrpc/message.go

// Request JSON-RPC 2.0 请求
type Request struct {
    JSONRPC string          `json:"jsonrpc"`          // 固定 "2.0"
    Method  string          `json:"method"`           // 方法名，如 "user.login"
    Params  json.RawMessage `json:"params,omitempty"` // 请求参数，与 HTTP body 一致
    ID      interface{}     `json:"id,omitempty"`     // 请求 ID (string 或 number)
}
```

### 1.2 JSON-RPC 2.0 响应格式

```go
// Response JSON-RPC 2.0 响应
type Response struct {
    JSONRPC string      `json:"jsonrpc"`          // 固定 "2.0"
    Result  interface{} `json:"result,omitempty"` // 成功时的结果，与 HTTP 响应一致
    Error   *Error      `json:"error,omitempty"`  // 错误时的信息
    ID      interface{} `json:"id"`               // 对应请求的 ID
}

// Error JSON-RPC 2.0 错误（扩展了 reason 字段）
type Error struct {
    Code    int               `json:"code"`           // JSON-RPC 错误码（负整数）
    Reason  string            `json:"reason"`         // 业务错误码，如 "Unauthenticated.PasswordIncorrect"
    Message string            `json:"message"`        // 错误描述
    Data    map[string]string `json:"data,omitempty"` // 额外上下文（对应 errorsx.Metadata）
}
```

### HTTP → JSON-RPC 错误码映射

| HTTP Status | JSON-RPC Code | 说明 |
|-------------|---------------|------|
| 400 | -32602 | Invalid params (标准) |
| 401 | -32001 | Unauthenticated |
| 403 | -32003 | Permission denied |
| 404 | -32004 | Resource not found |
| 409 | -32009 | Conflict |
| 429 | -32029 | Too many requests |
| 500 | -32603 | Internal error (标准) |
| 503 | -32053 | Service unavailable |

### 1.3 消息格式对照

| 协议 | 请求 | 成功响应 | 错误响应 |
|-----|------|---------|---------|
| **HTTP** | `POST /v1/user/login` | `{"access_token":"xxx"}` | `{"code":401,"reason":"xxx","message":"xxx","metadata":{...}}` |
| **gRPC** | `Login(LoginRequest)` | `LoginResponse` | `status.Error(code, msg)` |
| **WebSocket** | 见下方 | 见下方 | 见下方 |

**WebSocket 请求 (JSON-RPC 2.0)：**
```json
{
    "jsonrpc": "2.0",
    "method": "user.login",
    "params": {"username": "test", "password": "123456"},
    "id": 1
}
```
↑ `params` 与 HTTP body 完全一致

**WebSocket 成功响应：**
```json
{
    "jsonrpc": "2.0",
    "result": {"access_token": "xxx", "expires_at": 1234567890},
    "id": 1
}
```
↑ `result` 与 HTTP 响应完全一致

**WebSocket 错误响应：**
```json
{
    "jsonrpc": "2.0",
    "error": {
        "code": -32001,
        "reason": "Unauthenticated.PasswordIncorrect",
        "message": "Password is incorrect.",
        "data": {"field": "password"}
    },
    "id": 1
}
```
↑ `error.code` 为 JSON-RPC 错误码，`error.reason` 为业务错误码，`error.data` 携带额外上下文

**服务端推送 (Notification)：**
```json
{
    "jsonrpc": "2.0",
    "method": "order.created",
    "params": {"order_id": "123", "status": "pending"}
}
```
↑ 无 `id` 字段，客户端不需要响应

### 1.4 Go 实现

```go
// pkg/jsonrpc/message.go

package jsonrpc

import (
    "bingo/pkg/errorsx"
)

const Version = "2.0"

// NewResponse 创建成功响应
func NewResponse(id interface{}, result interface{}) *Response {
    return &Response{
        JSONRPC: Version,
        Result:  result,
        ID:      id,
    }
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(id interface{}, err error) *Response {
    e := errorsx.FromError(err)
    return &Response{
        JSONRPC: Version,
        Error: &Error{
            Code:    e.JSONRPCCode(), // HTTP code → JSON-RPC code
            Reason:  e.Reason,
            Message: e.Message,
            Data:    e.Metadata, // Metadata → Data
        },
        ID: id,
    }
}

// NewNotification 创建服务端推送（无 id）
func NewNotification(method string, params interface{}) *Response {
    return &Response{
        JSONRPC: Version,
        Result:  params, // Notification 使用 result 字段传递数据
    }
}
```

## 第二阶段：WebSocket 适配器

### 2.1 核心适配器

```go
// pkg/jsonrpc/adapter.go

package jsonrpc

import (
    "context"
    "encoding/json"
    "reflect"

    "google.golang.org/protobuf/proto"

    "bingo/internal/pkg/errno"
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
// method: JSON-RPC 方法名，如 "user.login"
// handler: Handler 方法
// reqType: Proto 请求类型的空实例，如 &pb.LoginRequest{}
func (a *Adapter) Register(method string, handler HandlerFunc, reqType proto.Message) {
    a.handlers[method] = &handlerInfo{
        handler:     handler,
        requestType: reflect.TypeOf(reqType).Elem(),
    }
}

// Handle 处理 JSON-RPC 请求
func (a *Adapter) Handle(ctx context.Context, req *Request) *Response {
    // 1. 查找 Handler
    info, ok := a.handlers[req.Method]
    if !ok {
        return NewErrorResponse(req.ID,
            errno.ErrNotFound.WithMessage("Method not found: %s", req.Method))
    }

    // 2. 创建 Proto 请求实例
    protoReq := reflect.New(info.requestType).Interface().(proto.Message)

    // 3. JSON → Proto
    if len(req.Params) > 0 {
        if err := json.Unmarshal(req.Params, protoReq); err != nil {
            return NewErrorResponse(req.ID,
                errno.ErrBind.WithMessage(err.Error()))
        }
    }

    // 4. 调用 Handler
    resp, err := info.handler(ctx, protoReq)
    if err != nil {
        return NewErrorResponse(req.ID, err)
    }

    // 5. Proto → JSON
    data, _ := json.Marshal(resp)
    var result interface{}
    json.Unmarshal(data, &result)

    return NewResponse(req.ID, result)
}
```

### 2.2 注册 Handlers

```go
// internal/apiserver/handler/ws/router.go

package ws

import (
    "context"

    "google.golang.org/protobuf/proto"

    "bingo/internal/apiserver/biz"
    "bingo/pkg/jsonrpc"
    pb "bingo/pkg/proto/apiserver/v1"
)

func RegisterHandlers(a *jsonrpc.Adapter, b biz.IBiz) {
    // 方法名建议使用 "domain.action" 格式
    a.Register("user.login", func(ctx context.Context, req proto.Message) (proto.Message, error) {
        return b.User().Login(ctx, req.(*pb.LoginRequest))
    }, &pb.LoginRequest{})

    a.Register("user.create", func(ctx context.Context, req proto.Message) (proto.Message, error) {
        return b.User().Create(ctx, req.(*pb.CreateUserRequest))
    }, &pb.CreateUserRequest{})

    a.Register("user.get", func(ctx context.Context, req proto.Message) (proto.Message, error) {
        return b.User().Get(ctx, req.(*pb.GetUserRequest))
    }, &pb.GetUserRequest{})

    a.Register("user.changePassword", func(ctx context.Context, req proto.Message) (proto.Message, error) {
        return b.User().ChangePassword(ctx, req.(*pb.ChangePasswordRequest))
    }, &pb.ChangePasswordRequest{})
}
```

### 2.3 简化注册（泛型版本）

```go
// pkg/jsonrpc/generic.go

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

// 使用示例
func RegisterHandlers(a *jsonrpc.Adapter, b biz.IBiz) {
    jsonrpc.Register(a, "user.login", b.User().Login, &pb.LoginRequest{})
    jsonrpc.Register(a, "user.create", b.User().Create, &pb.CreateUserRequest{})
    jsonrpc.Register(a, "user.get", b.User().Get, &pb.GetUserRequest{})
}
```

## 第三阶段：WebSocket 客户端改造

### 3.1 Client 结构体

```go
// pkg/ws/server/client.go

package server

import (
    "context"
    "sync"
    "time"

    "github.com/gorilla/websocket"

    "bingo/internal/pkg/contextx"
    "bingo/pkg/ws/adapter"
)

type Client struct {
    conn    *websocket.Conn
    send    chan []byte
    ctx     context.Context    // 携带认证信息的 context
    adapter *adapter.Adapter   // 消息处理适配器

    // 连接信息
    Addr          string
    AppID         uint32
    UserID        string
    FirstTime     uint64
    HeartbeatTime uint64
    LoginTime     uint64

    mu sync.RWMutex
}

func NewClient(conn *websocket.Conn, ctx context.Context, adapter *adapter.Adapter) *Client {
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

// Context 返回带认证信息的 context
func (c *Client) Context() context.Context {
    return c.ctx
}

// SetUserInfo 登录成功后设置用户信息
func (c *Client) SetUserInfo(userID string, appID uint32) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.UserID = userID
    c.AppID = appID
    c.LoginTime = uint64(time.Now().Unix())

    // 更新 context
    c.ctx = contextx.WithUserID(c.ctx, userID)
}
```

### 3.2 消息处理

```go
// pkg/ws/server/client.go

func (c *Client) readPump() {
    defer func() {
        ClientManager.Unregister <- c
        c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, msg, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Errorw("WebSocket read error", "err", err, "addr", c.Addr)
            }
            break
        }

        // 处理消息
        c.handleMessage(msg)
    }
}

func (c *Client) handleMessage(data []byte) {
    // 1. 解析 JSON-RPC 请求
    var req jsonrpc.Request
    if err := json.Unmarshal(data, &req); err != nil {
        resp := jsonrpc.NewErrorResponse(nil, errno.ErrBind.WithMessage("Invalid JSON"))
        c.SendJSON(resp)
        return
    }

    log.Infow("WebSocket request",
        "id", req.ID,
        "method", req.Method,
        "addr", c.Addr,
        "user_id", contextx.UserID(c.ctx),
    )

    // 2. 特殊方法处理（如心跳）
    if req.Method == "heartbeat" {
        c.Heartbeat()
        c.SendJSON(jsonrpc.NewResponse(req.ID, nil))
        return
    }

    // 3. 通过适配器调用 Biz 层
    resp := c.adapter.Handle(c.ctx, &req)

    // 4. 发送响应
    c.SendJSON(resp)

    // 5. 记录日志
    if resp.Error != nil {
        log.Infow("WebSocket response", "id", req.ID, "method", req.Method, "error", resp.Error.Reason)
    } else {
        log.Infow("WebSocket response", "id", req.ID, "method", req.Method, "success", true)
    }
}

func (c *Client) SendJSON(v interface{}) {
    data, err := json.Marshal(v)
    if err != nil {
        log.Errorw("JSON marshal error", "err", err)
        return
    }
    c.send <- data
}
```

## 第四阶段：连接认证

### 4.1 WebSocket 升级时认证

```go
// pkg/ws/server/server.go

package server

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"

    "bingo/internal/pkg/contextx"
    "bingo/internal/pkg/errno"
    "bingo/pkg/token"
    "bingo/pkg/ws/adapter"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type Server struct {
    adapter   *adapter.Adapter
    retriever UserRetriever
}

// UserRetriever 用户信息获取接口（与 gRPC 认证复用）
type UserRetriever interface {
    GetUser(ctx context.Context, userID string) (*contextx.UserInfo, error)
}

func NewServer(adapter *adapter.Adapter, retriever UserRetriever) *Server {
    return &Server{
        adapter:   adapter,
        retriever: retriever,
    }
}

// ServeWs 处理 WebSocket 连接
func (s *Server) ServeWs(c *gin.Context) {
    // 1. 获取 Token（支持 query 参数和 header）
    tokenStr := c.Query("token")
    if tokenStr == "" {
        tokenStr = extractBearerToken(c.GetHeader("Authorization"))
    }

    // 2. 创建 context
    ctx := context.Background()
    ctx = contextx.WithRequestID(ctx, c.GetHeader("X-Request-ID"))
    ctx = contextx.WithClientIP(ctx, c.ClientIP())

    // 3. 如果有 token，进行认证
    if tokenStr != "" {
        claims, err := token.Parse(tokenStr)
        if err != nil {
            c.JSON(401, gin.H{
                "code":    errno.ErrTokenInvalid.Reason,
                "message": errno.ErrTokenInvalid.Message,
            })
            return
        }

        userInfo, err := s.retriever.GetUser(ctx, claims.Subject)
        if err != nil {
            c.JSON(401, gin.H{
                "code":    errno.ErrUserNotFound.Reason,
                "message": errno.ErrUserNotFound.Message,
            })
            return
        }

        // 注入用户信息到 context
        ctx = contextx.WithUserID(ctx, userInfo.UID)
        ctx = contextx.WithUsername(ctx, userInfo.Username)
        ctx = contextx.WithUserInfo(ctx, userInfo)
    }

    // 4. 升级连接
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Errorw("WebSocket upgrade failed", "err", err)
        return
    }

    // 5. 创建客户端
    client := NewClient(conn, ctx, s.adapter)

    // 6. 如果已认证，直接标记为已登录
    if userID := contextx.UserID(ctx); userID != "" {
        client.SetUserInfo(userID, 0)
        ClientManager.Login <- client
    } else {
        ClientManager.Register <- client
    }

    // 7. 启动读写协程
    go client.writePump()
    go client.readPump()
}

func extractBearerToken(auth string) string {
    const prefix = "Bearer "
    if len(auth) > len(prefix) && auth[:len(prefix)] == prefix {
        return auth[len(prefix):]
    }
    return ""
}
```

### 4.2 连接后认证（user.login 方法）

对于不带 token 连接的客户端，可以通过 `user.login` 方法认证：

```go
// pkg/ws/server/login.go

// handleLogin 特殊处理 user.login 方法，因为需要更新 Client 状态
func (s *Server) handleLogin(client *Client, req *jsonrpc.Request) *jsonrpc.Response {
    // 1. 调用 Biz 层
    resp := s.adapter.Handle(client.ctx, req)

    // 2. 如果登录成功，更新 Client 状态
    if resp.Error == nil {
        // 从响应中提取 userID
        if result, ok := resp.Result.(map[string]interface{}); ok {
            if userID, ok := result["user_id"].(string); ok {
                client.SetUserInfo(userID, 0)
                ClientManager.Login <- client
            }
        }
    }

    return resp
}
```

## 第五阶段：推送消息

### 5.1 服务端推送

服务端推送使用 JSON-RPC 2.0 的 Notification 格式（无 `id` 字段）：

```go
// pkg/ws/server/push.go

package server

import (
    "encoding/json"

    "bingo/internal/pkg/errno"
    "bingo/pkg/jsonrpc"
)

// Push 向指定用户推送消息（Notification 格式）
func Push(userID string, method string, params interface{}) error {
    client := ClientManager.GetByUserID(userID)
    if client == nil {
        return errno.ErrUserNotFound
    }

    // Notification: 无 id 字段
    msg := jsonrpc.NewNotification(method, params)
    client.SendJSON(msg)
    return nil
}

// Broadcast 广播消息
func Broadcast(method string, params interface{}) {
    msg := jsonrpc.NewNotification(method, params)
    msgBytes, _ := json.Marshal(msg)
    ClientManager.Broadcast <- msgBytes
}

// PushToApp 向指定 App 的所有用户推送
func PushToApp(appID uint32, method string, params interface{}) {
    msg := jsonrpc.NewNotification(method, params)
    msgBytes, _ := json.Marshal(msg)
    ClientManager.BroadcastToApp(appID, msgBytes)
}
```

### 5.2 从 Biz 层触发推送

```go
// internal/apiserver/biz/order/order.go

package order

import (
    "context"

    "bingo/pkg/ws/server"
    pb "bingo/pkg/proto/apiserver/v1"
)

// Create 创建订单并推送通知
func (b *orderBiz) Create(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
    // ... 创建订单逻辑

    // 推送通知给用户（Notification 格式）
    server.Push(req.UserId, "order.created", map[string]interface{}{
        "order_id": order.ID,
        "status":   order.Status,
    })

    return &pb.CreateOrderResponse{OrderId: order.ID}, nil
}
```

## 第六阶段：完整启动流程

### 6.1 初始化代码

```go
// internal/apiserver/run.go

func run() error {
    // 1. 初始化依赖
    storeInstance := store.NewStore(db)
    bizInstance := biz.NewBiz(storeInstance)

    // 2. 创建 JSON-RPC 适配器并注册 handlers
    rpcAdapter := jsonrpc.NewAdapter()
    ws.RegisterHandlers(rpcAdapter, bizInstance)

    // 3. 创建 WebSocket Server
    userRetriever := NewUserRetriever(storeInstance)
    wsServer := server.NewServer(rpcAdapter, userRetriever)

    // 4. 启动 gRPC 服务器
    grpcHandler := grpc.NewHandler(bizInstance)
    grpcServer := NewGRPCServer(grpcHandler, userRetriever)
    go grpcServer.Run()

    // 5. 启动 gRPC-Gateway (HTTP)
    gwServer := NewGRPCGatewayServer(":8080", ":9090")
    go gwServer.Run()

    // 6. 启动 Gin（处理 WebSocket 和文件上传）
    g := bootstrap.InitGin()
    g.GET("/ws", wsServer.ServeWs)
    g.POST("/v1/file/upload", fileController.Upload)
    go g.Run(":8081")

    // 7. 启动 WebSocket ClientManager
    go server.ClientManager.Run()

    // 8. 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 9. 优雅关闭
    grpcServer.GracefulStop()
    gwServer.Shutdown(context.Background())

    return nil
}
```

### 6.2 端口分配

| 协议 | 端口 | 用途 |
|-----|------|-----|
| gRPC | 9090 | gRPC 直连（服务间调用） |
| HTTP | 8080 | gRPC-Gateway（对外 API） |
| HTTP | 8081 | Gin（WebSocket + 文件上传） |

或者合并 HTTP 端口：

```go
// 使用一个 HTTP 端口处理所有
mux := http.NewServeMux()
mux.Handle("/", gwmux)                    // gRPC-Gateway
mux.Handle("/ws", wsHandler)              // WebSocket
mux.Handle("/v1/file/upload", ginHandler) // 文件上传
```

## 消息格式对照

### 请求格式

**HTTP:**
```http
POST /v1/user/login
Content-Type: application/json

{"username": "test", "password": "123456"}
```

**gRPC:**
```protobuf
message LoginRequest {
    string username = 1;
    string password = 2;
}
```

**WebSocket (JSON-RPC 2.0):**
```json
{"jsonrpc": "2.0", "method": "user.login", "params": {"username": "test", "password": "123456"}, "id": 1}
```

### 成功响应

**HTTP:**
```json
{"access_token": "xxx", "expires_at": 1234567890}
```

**gRPC:**
```protobuf
message LoginResponse {
    string access_token = 1;
    int64 expires_at = 2;
}
```

**WebSocket (JSON-RPC 2.0):**
```json
{
    "jsonrpc": "2.0",
    "result": {"access_token": "xxx", "expires_at": 1234567890},
    "id": 1
}
```
↑ `result` 与 HTTP 响应完全一致

### 错误响应

**HTTP:**
```json
{"code": 401, "reason": "Unauthenticated.PasswordIncorrect", "message": "Password is incorrect.", "metadata": {"field": "password"}}
```

**gRPC:**
```
Status: UNAUTHENTICATED
Details: ErrorInfo{Reason: "Unauthenticated.PasswordIncorrect", Metadata: {"field": "password"}}
```

**WebSocket (JSON-RPC 2.0):**
```json
{
    "jsonrpc": "2.0",
    "error": {
        "code": -32001,
        "reason": "Unauthenticated.PasswordIncorrect",
        "message": "Password is incorrect.",
        "data": {"field": "password"}
    },
    "id": 1
}
```
↑ `error.code` 为 JSON-RPC 错误码，`error.data` 对应 HTTP/gRPC 的 `metadata`

## 迁移清单

### 需要新建的文件

1. **pkg/jsonrpc/message.go** - JSON-RPC 2.0 消息格式（Request/Response/Error）
2. **pkg/jsonrpc/adapter.go** - JSON-RPC → Biz 层适配器
3. **internal/apiserver/ws/router.go** - Handler 注册

### 需要修改的文件

1. **pkg/ws/server/client.go** - 添加 context 和 adapter
2. **pkg/ws/server/server.go** - 连接认证逻辑

### 需要删除的文件

1. **pkg/ws/common/errno.go** - 使用统一的 errno
2. **pkg/ws/common/response.go** - 使用 pkg/jsonrpc
3. **pkg/ws/message/** - 使用 pkg/jsonrpc
4. **internal/apiserver/ws/v1/** - 业务 Handler 移到 Biz 层

### 需要保留的文件

1. **pkg/ws/server/client_hub.go** - 连接管理
2. **pkg/ws/cache/*.go** - 用户在线状态缓存
3. **pkg/ws/task/*.go** - 定时任务

## 总结

### 核心设计原则

**不强求协议层完全统一，而是统一数据结构、错误格式和业务逻辑。**

### 统一后的收益

1. **业务逻辑只写一次** - Biz 层同时服务 HTTP/gRPC/WebSocket
2. **数据格式一致** - WebSocket 的 `params`/`result` 字段与 HTTP body 完全一致
3. **错误格式一致** - WebSocket 的 `error` 字段与 HTTP 错误格式一致
4. **认证机制复用** - WebSocket 使用相同的 token 验证逻辑
5. **易于维护** - 新增 API 只需修改 Proto、Biz 和注册一行代码
6. **类型安全** - 全链路使用 Proto 定义的类型
7. **行业标准** - JSON-RPC 2.0 有丰富的客户端库支持

### 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                Proto（数据定义 + 验证）                       │
└─────────────────────────────┬───────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
     ┌────────────┐    ┌────────────┐    ┌────────────┐
     │   HTTP     │    │   gRPC     │    │ WebSocket  │
     │ (Gateway)  │    │  Handler   │    │ (JSON-RPC) │
     └─────┬──────┘    └─────┬──────┘    └─────┬──────┘
           │                 │                 │
           └─────────────────┼─────────────────┘
                             ▼
                    ┌────────────────┐
                    │    Biz 层      │  ← 三种协议共享
                    └────────────────┘
```
