# WebSocket 设计与实现

本文档介绍 Bingo 的 WebSocket 实现，采用 JSON-RPC 2.0 协议，支持中间件、分组路由和连接管理。

## 目录

1. [设计目标](#设计目标)
2. [消息格式](#消息格式)
3. [中间件架构](#中间件架构)
4. [路由与 Handler](#路由与-handler)
5. [认证流程](#认证流程)
6. [连接管理](#连接管理)
7. [推送与订阅](#推送与订阅)
8. [适用场景](#适用场景)
9. [目录结构](#目录结构)
10. [未来规划](#未来规划)

---

## 设计目标

1. **统一消息格式** - 采用 JSON-RPC 2.0 标准，与 HTTP/gRPC 错误格式一致
2. **中间件模式** - 与 HTTP (Gin) 和 gRPC (interceptor) 保持一致的编程模型
3. **灵活分组** - 支持 public/private 分组，不同方法使用不同中间件链
4. **业务复用** - WebSocket Handler 直接调用 Biz 层，三协议共享业务逻辑

---

## 消息格式

WebSocket 采用 [JSON-RPC 2.0](https://www.jsonrpc.org/specification) 规范，这是行业广泛采用的标准（MCP、以太坊、VSCode LSP 等）。

### 请求格式

```json
{
    "jsonrpc": "2.0",
    "method": "auth.login",
    "params": {"username": "test", "password": "123456", "platform": "web"},
    "id": 1
}
```

| 字段 | 类型 | 说明 |
|-----|------|-----|
| jsonrpc | string | 固定 "2.0" |
| method | string | 方法名，如 "auth.login"、"subscribe" |
| params | object | 请求参数，与 HTTP body 格式一致 |
| id | string/number | 请求 ID，用于匹配响应 |

### 成功响应

```json
{
    "jsonrpc": "2.0",
    "result": {"accessToken": "xxx", "expiresAt": 1234567890},
    "id": 1
}
```

`result` 字段与 HTTP 响应 body 格式完全一致。

### 错误响应

```json
{
    "jsonrpc": "2.0",
    "error": {
        "code": -32001,
        "reason": "Unauthorized",
        "message": "Login required"
    },
    "id": 1
}
```

| 字段 | 说明 |
|-----|------|
| code | JSON-RPC 错误码（负整数） |
| reason | 业务错误码，与 HTTP 响应的 reason 一致 |
| message | 错误描述 |

### 服务端推送

服务端主动推送使用 Notification 格式（无 `id` 字段）：

```json
{
    "jsonrpc": "2.0",
    "method": "session.kicked",
    "params": {"reason": "您的账号已在其他设备登录"}
}
```

### HTTP → JSON-RPC 错误码映射

| HTTP Status | JSON-RPC Code | 说明 |
|-------------|---------------|------|
| 400 | -32602 | Invalid params |
| 401 | -32001 | Unauthorized |
| 403 | -32003 | Permission denied |
| 404 | -32004 | Not found |
| 429 | -32029 | Too many requests |
| 500 | -32603 | Internal error |

---

## 中间件架构

WebSocket 中间件与 HTTP/gRPC 保持一致的编程模型。

### 架构概览

```
WebSocket Message
      │
      ▼
┌─────────┐  ┌───────────┐  ┌───────────┐  ┌────────┐
│Recovery │→ │ RequestID │→ │  Logger   │→ │RateLimit│  ← 全局中间件
└─────────┘  └───────────┘  └───────────┘  └────────┘
      │
      ▼
┌─────────────────────────────────────────────────────┐
│                   路由分发                           │
│  ┌─────────────────┐    ┌─────────────────────┐     │
│  │   Public Group  │    │   Private Group     │     │
│  │                 │    │  ┌──────┐           │     │
│  │  • heartbeat    │    │  │ Auth │           │     │
│  │  • auth.login   │    │  └──────┘           │     │
│  │                 │    │  • subscribe        │     │
│  │                 │    │  • auth.user-info   │     │
│  └─────────────────┘    └─────────────────────┘     │
└─────────────────────────────────────────────────────┘
      │
      ▼
┌─────────┐
│ Handler │ → Biz Layer
└─────────┘
```

### 核心类型

```go
// pkg/ws/middleware.go

// Context 中间件上下文，嵌入 context.Context 可直接传递给 Biz 层
type Context struct {
    context.Context               // 嵌入标准 context，可直接传递给下层
    Request   *jsonrpc.Request    // JSON-RPC 请求
    Client    *Client             // WebSocket 客户端
    Method    string              // 方法名
    StartTime time.Time           // 请求开始时间
}

// 响应辅助方法
func (c *Context) JSON(data any) *jsonrpc.Response   // 返回成功响应
func (c *Context) Error(err error) *jsonrpc.Response // 返回错误响应

// Handler 消息处理函数
type Handler func(*Context) *jsonrpc.Response

// Middleware 中间件函数
type Middleware func(Handler) Handler

// Chain 组合多个中间件
func Chain(middlewares ...Middleware) Middleware
```

### 内置中间件

| 中间件 | 位置 | 说明 |
|-------|------|------|
| Recovery | `pkg/ws/middleware/recovery.go` | 捕获 panic，返回 500 错误 |
| RequestID | `pkg/ws/middleware/requestid.go` | 注入 request-id 到 context |
| Logger | `pkg/ws/middleware/logger.go` | 记录请求日志和延迟 |
| Auth | `pkg/ws/middleware/auth.go` | 验证用户已登录 |
| RateLimit | `pkg/ws/middleware/ratelimit.go` | 令牌桶限流 |
| LoginStateUpdater | `pkg/ws/middleware/login.go` | 登录成功后更新客户端状态 |

---

## 路由与 Handler

### Router

```go
// pkg/ws/router.go

// Router WebSocket 方法路由器
type Router struct {
    middlewares []Middleware
    handlers    map[string]*handlerEntry
}

// NewRouter 创建路由器
func NewRouter() *Router

// Use 添加全局中间件
func (r *Router) Use(middlewares ...Middleware) *Router

// Handle 注册方法处理器
func (r *Router) Handle(method string, handler Handler, middlewares ...Middleware)

// Group 创建分组
func (r *Router) Group(middlewares ...Middleware) *Group

// Dispatch 分发请求
func (r *Router) Dispatch(c *Context) *jsonrpc.Response
```

### Handler 注册示例

```go
// internal/apiserver/router/ws.go

func RegisterWSHandlers(router *ws.Router) {
    h := wshandler.NewHandler(store.S)

    // 全局中间件
    router.Use(
        middleware.Recovery,
        middleware.RequestID,
        middleware.Logger,
        middleware.RateLimit(&middleware.RateLimitConfig{
            Default: 10,
            Methods: map[string]float64{
                "heartbeat": 0, // 不限制
            },
        }),
    )

    // 公开方法（无需认证）
    public := router.Group()
    public.Handle("heartbeat", ws.HeartbeatHandler)
    public.Handle("system.healthz", h.Healthz)
    public.Handle("auth.login", h.Login, middleware.LoginStateUpdater)

    // 需要认证的方法
    private := router.Group(middleware.Auth)
    private.Handle("subscribe", ws.SubscribeHandler)
    private.Handle("unsubscribe", ws.UnsubscribeHandler)
    private.Handle("auth.user-info", h.UserInfo)
}
```

### Handler 签名

所有 Handler 使用统一的签名，类似 Gin 的 `func(c *gin.Context)`：

```go
// pkg/ws/middleware.go

// Handler 消息处理函数
type Handler func(*Context) *jsonrpc.Response

// Context 包含请求的所有信息，嵌入 context.Context 可直接传递给 Biz 层
type Context struct {
    context.Context               // 嵌入标准 context
    Request   *jsonrpc.Request    // JSON-RPC 请求
    Client    *Client             // WebSocket 客户端
    Method    string              // 方法名
    StartTime time.Time           // 请求开始时间
}

// BindParams 解析请求参数到结构体
func (c *Context) BindParams(v any) error

// BindValidate 解析并验证请求参数
func (c *Context) BindValidate(v any) error

// JSON 返回成功响应
func (c *Context) JSON(data any) *jsonrpc.Response

// Error 返回错误响应
func (c *Context) Error(err error) *jsonrpc.Response
```

### Handler 实现示例

```go
// internal/apiserver/handler/ws/auth.go

func (h *Handler) Login(c *ws.Context) *jsonrpc.Response {
    var req v1.LoginRequest
    if err := c.BindValidate(&req); err != nil {
        return c.Error(errno.ErrBind.SetMessage(err.Error()))
    }

    resp, err := h.b.Auth().Login(c, &req)  // ws.Context 嵌入 context.Context，可直接传递
    if err != nil {
        return c.Error(err)
    }

    return c.JSON(resp)
}

func (h *Handler) UserInfo(c *ws.Context) *jsonrpc.Response {
    uid := c.UserID()
    if uid == "" {
        return c.Error(errno.ErrTokenInvalid)
    }

    user, err := store.S.User().GetByUID(c, uid)
    if err != nil {
        return c.Error(errno.ErrUserNotFound)
    }

    return c.JSON(&v1.UserInfo{...})
}
```

### 内置 Handler

| Handler | 说明 |
|---------|------|
| HeartbeatHandler | 心跳响应，返回服务器时间 |
| SubscribeHandler | 订阅主题 |
| UnsubscribeHandler | 取消订阅 |

---

## 认证流程

WebSocket 支持两种认证方式：

### 方式一：连接后登录

1. 客户端建立 WebSocket 连接（匿名状态）
2. 发送 `auth.login` 请求
3. 服务端验证成功后，`LoginStateUpdater` 中间件更新客户端状态
4. 后续请求可访问需要认证的方法

```
Client                          Server
  │                                │
  │──── WebSocket Connect ────────>│  匿名连接
  │                                │
  │──── auth.login ───────────────>│
  │<─── {accessToken: xxx} ────────│  登录成功
  │                                │  LoginStateUpdater 更新状态
  │                                │
  │──── subscribe ────────────────>│  已认证，允许
  │<─── {subscribed: [...]} ───────│
```

### 方式二：连接时携带 Token（可选扩展）

可以在 WebSocket 升级时通过 query 参数或 header 传递 token：

```
ws://example.com/ws?token=xxx
```

### 认证状态管理

```go
// pkg/ws/client.go

// IsAuthenticated 检查是否已登录
func (c *Client) IsAuthenticated() bool {
    return c.UserID != "" && c.Platform != "" && c.LoginTime > 0
}

// NotifyLogin 通知 Hub 登录成功
func (c *Client) NotifyLogin(userID, platform string, tokenExpiresAt int64)
```

---

## 连接管理

### Hub

Hub 是 WebSocket 连接的中央管理器，负责：
- 客户端注册/注销
- 用户登录状态管理
- 主题订阅/发布
- 连接清理

```go
// pkg/ws/hub.go

type Hub struct {
    anonymous   map[*Client]bool  // 匿名连接
    clients     map[*Client]bool  // 已认证连接
    users       map[string]*Client // platform_userID -> Client
    clientsByID map[string]*Client // clientID -> Client
    topics      map[string]map[*Client]bool // 主题订阅
}

// 管理 API
func (h *Hub) GetClient(clientID string) *Client
func (h *Hub) GetClientsByUser(userID string) []*Client
func (h *Hub) KickClient(clientID string, reason string) bool
func (h *Hub) KickUser(userID string, reason string) int
func (h *Hub) Stats() *HubStats

// 推送 API
func (h *Hub) PushToTopic(topic, method string, data any)
func (h *Hub) PushToUser(platform, userID, method string, data any)
func (h *Hub) PushToUserAllPlatforms(userID, method string, data any)
```

### 连接生命周期

```
┌─────────────────────────────────────────────────────────────┐
│                      连接生命周期                            │
│                                                             │
│  Connect ──> Register ──> [Anonymous]                       │
│                              │                              │
│                         auth.login                          │
│                              │                              │
│                              ▼                              │
│                        [Authenticated]                      │
│                              │                              │
│            ┌─────────────────┼─────────────────┐            │
│            ▼                 ▼                 ▼            │
│      Token 过期         心跳超时          被踢下线           │
│            │                 │                 │            │
│            └─────────────────┼─────────────────┘            │
│                              ▼                              │
│                         Unregister                          │
└─────────────────────────────────────────────────────────────┘
```

### 单设备登录

同一用户在同一平台（web/mobile/desktop）只能有一个活跃连接。新连接登录时，旧连接会收到踢出通知：

```json
{
    "jsonrpc": "2.0",
    "method": "session.kicked",
    "params": {"reason": "您的账号已在其他设备登录"}
}
```

### 心跳机制

采用双层心跳架构：

```
┌─────────────────────────────────────────────────────────────┐
│                    协议层 (WebSocket)                        │
│                                                              │
│  服务端 ──── ping (每54s) ────→ 客户端                       │
│  服务端 ←─── pong ─────────── 客户端                        │
│                                                              │
│  目的：检测 TCP 连接活性                                      │
│  超时：60s 未收到 pong → 断开                                │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    应用层 (JSON-RPC)                         │
│                                                              │
│  服务端 ←─── heartbeat (每30s) ─── 客户端                   │
│  服务端 ──── response ──────────→ 客户端                    │
│                                                              │
│  目的：确认客户端还在消费数据、保持 NAT 映射                   │
│  超时：90s 未收到任何消息 → 断开                             │
└─────────────────────────────────────────────────────────────┘
```

**为什么需要双层心跳？**

| 层级 | 检测目标 | 场景 |
|-----|---------|------|
| 协议层 | TCP 连接是否存活 | 网络断开、客户端崩溃 |
| 应用层 | 客户端是否在消费数据 | App 被杀但 TCP 未断 |

### 客户端要求

1. 登录后 **每 30 秒** 发送一次 `heartbeat` 请求
2. **90 秒** 内无任何消息将被服务端断开
3. 断开后需自行实现 **重连逻辑**
4. 重连后需 **重新登录** 并 **重新订阅** Topic

---

## 推送与订阅

### 推送 API

```go
// 用户推送
hub.PushToUser("ios", "user123", "order.created", data)
hub.PushToUserAllPlatforms("user123", "security.alert", data)

// Topic 推送
hub.PushToTopic("group:123", "message.new", data)

// 广播
hub.Broadcast("system.maintenance", data)
```

### Topic 订阅

Topic 用于发布/订阅模式，支持多种实时数据场景：

| 前缀 | 用途 | 示例 |
|-----|------|------|
| `group:` | 群聊 | `group:123` |
| `room:` | 聊天室 | `room:lobby` |
| `ticker:` | 实时行情 | `ticker:BTC/USDT` |
| `device:` | IoT 设备 | `device:12345` |

**订阅请求**：
```json
{
    "jsonrpc": "2.0",
    "method": "subscribe",
    "params": {"topics": ["group:123", "room:lobby"]},
    "id": 2
}
```

**订阅响应**：
```json
{
    "jsonrpc": "2.0",
    "result": {"subscribed": ["group:123", "room:lobby"]},
    "id": 2
}
```

---

## 适用场景

| 场景 | Topic 示例 | 消息特点 |
|-----|-----------|---------|
| 即时通讯 | `group:{groupID}` | 广播、@提醒 |
| 协同文档 | `doc:{docID}` | 多人编辑、光标同步 |
| 实时行情 | `ticker:{symbol}` | 高频小消息 |
| 订单通知 | 用户私有推送 | 状态变更 |
| 系统维护 | 广播 | 全局通知 |

---

## 目录结构

```
pkg/ws/
├── client.go           # 客户端连接
├── hub.go              # 连接管理
├── hub_config.go       # Hub 配置
├── router.go           # 路由器和分组
├── middleware.go       # 中间件类型定义
├── handlers.go         # 内置 Handler
├── platform.go         # 平台常量
└── middleware/         # 内置中间件
    ├── recovery.go
    ├── requestid.go
    ├── logger.go
    ├── auth.go
    ├── ratelimit.go
    └── login.go

pkg/jsonrpc/
├── message.go          # Request/Response 类型
└── response.go         # 响应构造函数

internal/apiserver/
├── ws.go               # WebSocket 初始化
├── router/
│   └── ws.go           # Handler 注册
└── handler/ws/
    ├── handler.go      # Handler 定义
    ├── auth.go         # 认证相关
    └── system.go       # 系统方法
```

---

## 未来规划

> 以下功能已设计但尚未实现

### Metrics 中间件

集成 Prometheus 监控：

```go
// 计划实现
var (
    requestsTotal = prometheus.NewCounterVec(...)
    requestDuration = prometheus.NewHistogramVec(...)
)

func Metrics(next Handler) Handler {
    // 记录请求数和延迟
}
```

### 连接数限制

```go
// 计划实现
type HubConfig struct {
    MaxConnectionsPerUser int  // 单用户最大连接数
}
```

---

## 相关文档

- [可插拔协议层](protocol-layer.md) - HTTP/gRPC/WebSocket 统一架构
- [gRPC-Gateway 集成](grpc-gateway.md) - Gateway 模式配置与使用
- [统一错误处理](unified-error-handling.md) - 三协议错误格式统一

---

**下一步**：了解 [gRPC-Gateway 集成](grpc-gateway.md)，学习如何让一份代码同时支持 HTTP 和 gRPC。
