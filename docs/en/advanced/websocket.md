# WebSocket Design and Implementation

> **Standalone Library Available**: The WebSocket module is now available as a standalone library at [github.com/bingo-project/websocket](https://github.com/bingo-project/websocket). You can use it independently without the full Bingo framework.

This document describes Bingo's WebSocket implementation, using JSON-RPC 2.0 protocol with middleware, grouped routing, and connection management support.

## Table of Contents

1. [Design Goals](#design-goals)
2. [Message Format](#message-format)
3. [Middleware Architecture](#middleware-architecture)
4. [Routing and Handler](#routing-and-handler)
5. [Authentication Flow](#authentication-flow)
6. [Connection Management](#connection-management)
7. [Push and Subscribe](#push-and-subscribe)
8. [Prometheus Metrics](#prometheus-metrics)
9. [Connection Limits](#connection-limits)
10. [Use Cases](#use-cases)
11. [Directory Structure](#directory-structure)

---

## Design Goals

1. **Unified Message Format** - Use JSON-RPC 2.0 standard, consistent error format with HTTP/gRPC
2. **Middleware Pattern** - Consistent programming model with HTTP (Gin) and gRPC (interceptor)
3. **Flexible Grouping** - Support public/private groups, different methods use different middleware chains
4. **Business Reuse** - WebSocket Handler directly calls Biz layer, all protocols share business logic

---

## Message Format

WebSocket uses [JSON-RPC 2.0](https://www.jsonrpc.org/specification) specification, a widely adopted industry standard (MCP, Ethereum, VSCode LSP, etc.).

### Request Format

```json
{
    "jsonrpc": "2.0",
    "method": "auth.login",
    "params": {"username": "test", "password": "123456", "platform": "web"},
    "id": 1
}
```

| Field | Type | Description |
|-------|------|-------------|
| jsonrpc | string | Fixed "2.0" |
| method | string | Method name, e.g., "auth.login", "subscribe" |
| params | object | Request parameters, same format as HTTP body |
| id | string/number | Request ID for matching responses |

### Success Response

```json
{
    "jsonrpc": "2.0",
    "result": {"accessToken": "xxx", "expiresAt": 1234567890},
    "id": 1
}
```

The `result` field format is identical to HTTP response body.

### Error Response

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

| Field | Description |
|-------|-------------|
| code | JSON-RPC error code (negative integer) |
| reason | Business error code, same as HTTP response reason |
| message | Error description |

### Server Push

Server-initiated push uses Notification format (no `id` field):

```json
{
    "jsonrpc": "2.0",
    "method": "session.kicked",
    "params": {"reason": "Your account has logged in on another device"}
}
```

### HTTP → JSON-RPC Error Code Mapping

| HTTP Status | JSON-RPC Code | Description |
|-------------|---------------|-------------|
| 400 | -32602 | Invalid params |
| 401 | -32001 | Unauthorized |
| 403 | -32003 | Permission denied |
| 404 | -32004 | Not found |
| 429 | -32029 | Too many requests |
| 500 | -32603 | Internal error |

---

## Middleware Architecture

WebSocket middleware maintains consistent programming model with HTTP/gRPC.

### Architecture Overview

```
WebSocket Message
      │
      ▼
┌─────────┐  ┌───────────┐  ┌───────────┐  ┌────────┐
│Recovery │→ │ RequestID │→ │  Logger   │→ │RateLimit│  ← Global middleware
└─────────┘  └───────────┘  └───────────┘  └────────┘
      │
      ▼
┌─────────────────────────────────────────────────────┐
│                   Route Dispatch                     │
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

### Core Types

```go
// github.com/bingo-project/websocket/context.go

// Context middleware context, embeds context.Context for direct Biz layer passing
type Context struct {
    context.Context               // Embedded standard context, can be passed directly to lower layers
    Request   *jsonrpc.Request    // JSON-RPC request
    Client    *Client             // WebSocket client
    Method    string              // Method name
    StartTime time.Time           // Request start time
}

// Response helper methods
func (c *Context) JSON(data any) *jsonrpc.Response   // Return success response
func (c *Context) Error(err error) *jsonrpc.Response // Return error response

// Handler message processing function
type Handler func(*Context) *jsonrpc.Response

// Middleware function
type Middleware func(Handler) Handler

// Chain combines multiple middlewares
func Chain(middlewares ...Middleware) Middleware
```

### Built-in Middleware

| Middleware | Location | Description |
|------------|----------|-------------|
| Recovery / RecoveryWithLogger | `middleware/recovery.go` | Catch panic, return 500 error; WithLogger version supports custom logger |
| RequestID | `middleware/requestid.go` | Inject request-id into context |
| Logger / LoggerWithLogger | `middleware/logger.go` | Log requests and latency; WithLogger version supports custom logger |
| Auth | `middleware/auth.go` | Verify user is logged in |
| RateLimitWithStore | `middleware/ratelimit.go` | Token bucket rate limiting with Redis storage |
| LoginStateUpdater | `middleware/login.go` | Update client state after successful login |

**Custom Logger:**

```go
// Using custom logger
import "github.com/marmotedu/iam/pkg/log"

router.Use(
    middleware.RecoveryWithLogger(log.L()),  // Use project logger
    middleware.LoggerWithLogger(log.L()),    // Use project logger
)
```

---

## Routing and Handler

### Router

```go
// github.com/bingo-project/websocket/router.go

// Router WebSocket method router
type Router struct {
    middlewares []Middleware
    handlers    map[string]*handlerEntry
}

// NewRouter creates a router
func NewRouter() *Router

// Use adds global middleware
func (r *Router) Use(middlewares ...Middleware) *Router

// Handle registers method handler
func (r *Router) Handle(method string, handler Handler, middlewares ...Middleware)

// Group creates a group
func (r *Router) Group(middlewares ...Middleware) *Group

// Dispatch dispatches request
func (r *Router) Dispatch(c *Context) *jsonrpc.Response
```

### Handler Registration Example

```go
// internal/apiserver/router/ws.go

func RegisterWSHandlers(router *ws.Router, rateLimitStore *middleware.RateLimiterStore) {
    h := wshandler.NewHandler(store.S)

    // Global middleware
    router.Use(
        middleware.Recovery,
        middleware.RequestID,
        middleware.Logger,
        middleware.RateLimitWithStore(rateLimitStore, &middleware.RateLimitConfig{
            Default: 10,
            Methods: map[string]float64{
                "heartbeat": 0, // No limit
            },
        }),
    )

    // Public methods (no authentication required)
    public := router.Group()
    public.Handle("heartbeat", ws.HeartbeatHandler)
    public.Handle("system.healthz", h.Healthz)
    public.Handle("auth.login", h.Login, middleware.LoginStateUpdater)

    // Methods requiring authentication
    private := router.Group(middleware.Auth)
    private.Handle("subscribe", ws.SubscribeHandler)
    private.Handle("unsubscribe", ws.UnsubscribeHandler)
    private.Handle("auth.user-info", h.UserInfo)
}
```

### Handler Signature

All Handlers use a unified signature, similar to Gin's `func(c *gin.Context)`:

```go
// github.com/bingo-project/websocket/context.go

// Handler message processing function
type Handler func(*Context) *jsonrpc.Response

// Context contains all request information, embeds context.Context for direct Biz layer passing
type Context struct {
    context.Context               // Embedded standard context
    Request   *jsonrpc.Request    // JSON-RPC request
    Client    *Client             // WebSocket client
    Method    string              // Method name
    StartTime time.Time           // Request start time
}

// BindParams parses request parameters into struct
func (c *Context) BindParams(v any) error

// BindValidate parses and validates request parameters
func (c *Context) BindValidate(v any) error

// JSON returns success response
func (c *Context) JSON(data any) *jsonrpc.Response

// Error returns error response
func (c *Context) Error(err error) *jsonrpc.Response
```

### Handler Implementation Example

```go
// internal/apiserver/handler/ws/auth.go

func (h *Handler) Login(c *ws.Context) *jsonrpc.Response {
    var req v1.LoginRequest
    if err := c.BindValidate(&req); err != nil {
        return c.Error(errno.ErrInvalidArgument.WithMessage(err.Error()))
    }

    resp, err := h.b.Auth().Login(c, &req)  // ws.Context embeds context.Context, can be passed directly
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

### Built-in Handlers

| Handler | Description |
|---------|-------------|
| HeartbeatHandler | Heartbeat response, returns server time |
| SubscribeHandler | Subscribe to topic |
| UnsubscribeHandler | Unsubscribe from topic |

---

## Authentication Flow

WebSocket supports two authentication methods:

### Method 1: Login After Connection

1. Client establishes WebSocket connection (anonymous state)
2. Send `auth.login` request
3. After server verification, `LoginStateUpdater` middleware updates client state
4. Subsequent requests can access authenticated methods

```
Client                          Server
  │                                │
  │──── WebSocket Connect ────────>│  Anonymous connection
  │                                │
  │──── auth.login ───────────────>│
  │<─── {accessToken: xxx} ────────│  Login successful
  │                                │  LoginStateUpdater updates state
  │                                │
  │──── subscribe ────────────────>│  Authenticated, allowed
  │<─── {subscribed: [...]} ───────│
```

### Method 2: Token in Connection (Optional Extension)

Token can be passed via query parameter or header during WebSocket upgrade:

```
ws://example.com/ws?token=xxx
```

### Authentication State Management

```go
// github.com/bingo-project/websocket/client.go

// IsAuthenticated checks if logged in
func (c *Client) IsAuthenticated() bool {
    return c.UserID != "" && c.Platform != "" && c.LoginTime > 0
}

// NotifyLogin notifies Hub of successful login
func (c *Client) NotifyLogin(userID, platform string, tokenExpiresAt int64)
```

---

## Connection Management

### Hub

Hub is the central manager for WebSocket connections, responsible for:
- Client registration/unregistration
- User login state management
- Topic subscription/publishing
- Connection cleanup

```go
// github.com/bingo-project/websocket/hub.go

type Hub struct {
    anonymous   map[*Client]bool  // Anonymous connections
    clients     map[*Client]bool  // Authenticated connections
    users       map[string]*Client // platform_userID -> Client
    clientsByID map[string]*Client // clientID -> Client
    topics      map[string]map[*Client]bool // Topic subscriptions
}

// Management API
func (h *Hub) GetClient(clientID string) *Client
func (h *Hub) GetClientsByUser(userID string) []*Client
func (h *Hub) KickClient(clientID string, reason string) bool
func (h *Hub) KickUser(userID string, reason string) int
func (h *Hub) Stats() *HubStats

// Push API
func (h *Hub) PushToTopic(topic, method string, data any)
func (h *Hub) PushToUser(platform, userID, method string, data any)
func (h *Hub) PushToUserAllPlatforms(userID, method string, data any)
```

### Connection Lifecycle

```
┌─────────────────────────────────────────────────────────────┐
│                      Connection Lifecycle                    │
│                                                              │
│  Connect ──> Register ──> [Anonymous]                        │
│                              │                               │
│                         auth.login                           │
│                              │                               │
│                              ▼                               │
│                        [Authenticated]                       │
│                              │                               │
│            ┌─────────────────┼─────────────────┐             │
│            ▼                 ▼                 ▼             │
│      Token Expired     Heartbeat Timeout    Kicked          │
│            │                 │                 │             │
│            └─────────────────┼─────────────────┘             │
│                              ▼                               │
│                         Unregister                           │
└─────────────────────────────────────────────────────────────┘
```

### Single Device Login

Only one active connection per user per platform (web/mobile/desktop). When a new connection logs in, the old connection receives a kick notification:

```json
{
    "jsonrpc": "2.0",
    "method": "session.kicked",
    "params": {"reason": "Your account has logged in on another device"}
}
```

### Heartbeat Mechanism

Dual-layer heartbeat architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                    Protocol Layer (WebSocket)                │
│                                                              │
│  Server ──── ping (every 54s) ────→ Client                  │
│  Server ←─── pong ─────────────── Client                    │
│                                                              │
│  Purpose: Detect TCP connection liveness                     │
│  Timeout: 60s without pong → disconnect                      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Application Layer (JSON-RPC)              │
│                                                              │
│  Server ←─── heartbeat (every 30s) ─── Client               │
│  Server ──── response ──────────────→ Client                │
│                                                              │
│  Purpose: Confirm client is consuming data, maintain NAT     │
│  Timeout: 90s without any message → disconnect              │
└─────────────────────────────────────────────────────────────┘
```

**Why Dual-Layer Heartbeat?**

| Layer | Detection Target | Scenario |
|-------|------------------|----------|
| Protocol | TCP connection alive | Network disconnect, client crash |
| Application | Client consuming data | App killed but TCP not closed |

### Client Requirements

1. After login, send `heartbeat` request **every 30 seconds**
2. **90 seconds** without any message will be disconnected by server
3. Implement **reconnection logic** after disconnect
4. After reconnection, **re-login** and **re-subscribe** to topics

---

## Push and Subscribe

### Push API

```go
// User push
hub.PushToUser("ios", "user123", "order.created", data)
hub.PushToUserAllPlatforms("user123", "security.alert", data)

// Topic push
hub.PushToTopic("group:123", "message.new", data)

// Broadcast
hub.Broadcast("system.maintenance", data)
```

### Topic Subscription

Topics are used for publish/subscribe pattern, supporting various real-time data scenarios:

| Prefix | Purpose | Example |
|--------|---------|---------|
| `group:` | Group chat | `group:123` |
| `room:` | Chat room | `room:lobby` |
| `ticker:` | Real-time quotes | `ticker:BTC/USDT` |
| `device:` | IoT device | `device:12345` |

**Subscribe Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "subscribe",
    "params": {"topics": ["group:123", "room:lobby"]},
    "id": 2
}
```

**Subscribe Response:**
```json
{
    "jsonrpc": "2.0",
    "result": {"subscribed": ["group:123", "room:lobby"]},
    "id": 2
}
```

---

## Prometheus Metrics

The WebSocket library includes built-in Prometheus metrics support for monitoring connection status and message throughput.

### Enable Metrics

```go
import "github.com/prometheus/client_golang/prometheus"

// Create and register metrics
metrics := websocket.NewMetrics("myapp", "websocket")
metrics.MustRegister(prometheus.DefaultRegisterer)

// Attach metrics to Hub
hub := websocket.NewHub(websocket.WithMetrics(metrics))
```

### Available Metrics

| Metric Name | Type | Description |
|-------------|------|-------------|
| `{namespace}_{subsystem}_connections_total` | Counter | Total connections |
| `{namespace}_{subsystem}_connections_current` | Gauge | Current connections |
| `{namespace}_{subsystem}_connections_authenticated` | Gauge | Authenticated connections |
| `{namespace}_{subsystem}_connections_anonymous` | Gauge | Anonymous connections |
| `{namespace}_{subsystem}_messages_sent_total` | Counter | Total messages sent |
| `{namespace}_{subsystem}_broadcasts_total` | Counter | Total broadcasts |
| `{namespace}_{subsystem}_errors_total` | Counter | Total errors (grouped by type) |
| `{namespace}_{subsystem}_topics_current` | Gauge | Current topics |
| `{namespace}_{subsystem}_subscriptions_total` | Counter | Total subscriptions |

---

## Connection Limits

The WebSocket library supports configuring maximum connections and per-user connection limits to prevent resource exhaustion.

### Configuration

```go
cfg := &websocket.HubConfig{
    MaxConnections: 10000,  // Max total connections (0 = unlimited)
    MaxConnsPerUser: 5,     // Max connections per user (0 = unlimited)
    // ... other config
}

hub := websocket.NewHubWithConfig(cfg)
```

### Usage

```go
// Check before accepting connection (optional, for early rejection)
if !hub.CanAcceptConnection() {
    http.Error(w, "Too many connections", http.StatusServiceUnavailable)
    return
}

// Check before login (optional)
if !hub.CanUserConnect(userID) {
    return c.Error(errors.New(429, "TooManyConnections", "Max connections reached"))
}

// Limits are also enforced automatically within Hub
```

---

## Use Cases

| Scenario | Topic Example | Message Characteristics |
|----------|---------------|------------------------|
| Instant Messaging | `group:{groupID}` | Broadcast, @mentions |
| Collaborative Docs | `doc:{docID}` | Multi-user editing, cursor sync |
| Real-time Quotes | `ticker:{symbol}` | High-frequency small messages |
| Order Notifications | User private push | Status changes |
| System Maintenance | Broadcast | Global notifications |

---

## Directory Structure

The WebSocket module is now a standalone library. Import paths:

```go
import (
    "github.com/bingo-project/websocket"          // Core types (Hub, Client, Router, Context)
    "github.com/bingo-project/websocket/jsonrpc"  // JSON-RPC message types
    "github.com/bingo-project/websocket/middleware" // Built-in middleware
)
```

Bingo integration files:

```
internal/apiserver/
├── ws.go               # WebSocket initialization
├── router/
│   └── ws.go           # Handler registration
└── handler/ws/
    ├── handler.go      # Handler definition
    ├── auth.go         # Authentication related
    └── system.go       # System methods
```

---

## Related Documentation

- [Pluggable Protocol Layer](protocol-layer.md) - HTTP/gRPC/WebSocket unified architecture
- [gRPC-Gateway Integration](grpc-gateway.md) - Gateway mode configuration and usage
- [Unified Error Handling](unified-error-handling.md) - Consistent error format across all protocols

---

**Next Step**: Learn about [gRPC-Gateway Integration](grpc-gateway.md) to understand how to support both HTTP and gRPC with a single codebase.
