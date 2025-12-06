# WebSocket 中间件设计

## 背景

当前 WebSocket 实现缺少像 HTTP/gRPC 那样的中间件机制。横切关注点（request-id、日志、认证、限流）硬编码在 `handleMessage()` 中，难以复用和扩展。

## 设计目标

1. **统一模式** - 与 HTTP (Gin middleware) 和 gRPC (interceptor) 保持一致的编程模型
2. **灵活组合** - 支持分组，不同方法可使用不同中间件链
3. **高性能** - 中间件开销 < 1μs，对业务无感知
4. **可观测** - 内置 metrics 和日志支持

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                     消息处理流程                              │
│                                                              │
│  WebSocket Message                                           │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────┐  ┌───────────┐  ┌───────────┐  ┌─────────┐     │
│  │Recovery │→ │ RequestID │→ │ RateLimit │→ │ Metrics │     │
│  └─────────┘  └───────────┘  └───────────┘  └─────────┘     │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                   路由分发                           │    │
│  │  ┌─────────────────┐    ┌─────────────────────┐     │    │
│  │  │   Public Group  │    │   Private Group     │     │    │
│  │  │                 │    │  ┌──────┐ ┌──────┐  │     │    │
│  │  │  • heartbeat    │    │  │ Auth │→│Logger│  │     │    │
│  │  │  • auth.login   │    │  └──────┘ └──────┘  │     │    │
│  │  │                 │    │  • subscribe        │     │    │
│  │  │                 │    │  • user.query       │     │    │
│  │  └─────────────────┘    └─────────────────────┘     │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────┐                                                 │
│  │ Handler │ → Biz Layer                                     │
│  └─────────┘                                                 │
└─────────────────────────────────────────────────────────────┘
```

## 核心类型定义

### MiddlewareContext

```go
// pkg/ws/middleware.go

// MiddlewareContext 中间件上下文，包含请求处理所需的所有信息
type MiddlewareContext struct {
    Ctx       context.Context   // 请求上下文，携带 request-id、user-id 等
    Request   *jsonrpc.Request  // JSON-RPC 请求
    Client    *Client           // WebSocket 客户端连接
    Method    string            // 方法名
    StartTime time.Time         // 请求开始时间（用于 metrics）
}

// UserID 从上下文获取用户 ID
func (mc *MiddlewareContext) UserID() string {
    return contextx.UserID(mc.Ctx)
}

// RequestID 从上下文获取请求 ID
func (mc *MiddlewareContext) RequestID() string {
    return contextx.RequestID(mc.Ctx)
}
```

### Handler 和 Middleware

```go
// Handler 消息处理函数
type Handler func(*MiddlewareContext) *jsonrpc.Response

// Middleware 中间件函数
type Middleware func(Handler) Handler
```

### Router（路由器）

```go
// pkg/ws/router.go

// Router WebSocket 方法路由器
type Router struct {
    middlewares []Middleware
    handlers    map[string]*handlerEntry
    groups      []*Group
}

type handlerEntry struct {
    handler     Handler
    middlewares []Middleware
}

// NewRouter 创建路由器
func NewRouter() *Router

// Use 添加全局中间件
func (r *Router) Use(middlewares ...Middleware) *Router

// Handle 注册方法处理器
func (r *Router) Handle(method string, handler Handler, middlewares ...Middleware)

// Group 创建方法分组
func (r *Router) Group(middlewares ...Middleware) *Group

// Dispatch 分发消息到对应处理器
func (r *Router) Dispatch(mc *MiddlewareContext) *jsonrpc.Response
```

### Group（分组）

```go
// Group 方法分组，共享中间件
type Group struct {
    router      *Router
    middlewares []Middleware
}

// Handle 在分组中注册方法
func (g *Group) Handle(method string, handler Handler, middlewares ...Middleware)

// Use 添加分组中间件
func (g *Group) Use(middlewares ...Middleware) *Group
```

## 内置中间件

### Recovery

```go
// pkg/ws/middleware/recovery.go

func Recovery(next Handler) Handler {
    return func(mc *MiddlewareContext) (resp *jsonrpc.Response) {
        defer func() {
            if r := recover(); r != nil {
                log.C(mc.Ctx).Errorw("WebSocket panic recovered",
                    "method", mc.Method,
                    "panic", r,
                    "stack", string(debug.Stack()),
                )
                resp = jsonrpc.NewErrorResponse(mc.Request.ID,
                    errorsx.New(500, "InternalError", "Internal server error"))
            }
        }()
        return next(mc)
    }
}
```

### RequestID

```go
// pkg/ws/middleware/requestid.go

func RequestID(next Handler) Handler {
    return func(mc *MiddlewareContext) *jsonrpc.Response {
        // 优先使用客户端提供的 request id
        requestID := ""
        if mc.Request.ID != nil {
            requestID = fmt.Sprintf("%v", mc.Request.ID)
        }
        if requestID == "" {
            requestID = uuid.New().String()
        }

        mc.Ctx = contextx.WithRequestID(mc.Ctx, requestID)
        return next(mc)
    }
}
```

### Logger

```go
// pkg/ws/middleware/logger.go

func Logger(next Handler) Handler {
    return func(mc *MiddlewareContext) *jsonrpc.Response {
        resp := next(mc)

        fields := []any{
            "method", mc.Method,
            "client_id", mc.Client.ID,
            "latency", time.Since(mc.StartTime),
        }

        if resp.Error != nil {
            fields = append(fields, "error", resp.Error.Reason)
            log.C(mc.Ctx).Warnw("WebSocket request failed", fields...)
        } else {
            log.C(mc.Ctx).Infow("WebSocket request", fields...)
        }

        return resp
    }
}
```

### Auth

```go
// pkg/ws/middleware/auth.go

func Auth(next Handler) Handler {
    return func(mc *MiddlewareContext) *jsonrpc.Response {
        if !mc.Client.IsAuthenticated() {
            return jsonrpc.NewErrorResponse(mc.Request.ID,
                errorsx.New(401, "Unauthorized", "Login required"))
        }

        // 将 user info 放入 context
        mc.Ctx = contextx.WithUserID(mc.Ctx, mc.Client.UserID)

        return next(mc)
    }
}
```

### RateLimit

```go
// pkg/ws/middleware/ratelimit.go

// RateLimitConfig 限流配置
type RateLimitConfig struct {
    Default      rate.Limit            // 默认限制（每秒请求数）
    Methods      map[string]rate.Limit // 按方法配置（0 = 不限制）
    BanThreshold int                   // 连续超限次数达到此值断开连接
}

func RateLimit(cfg *RateLimitConfig) Middleware {
    return func(next Handler) Handler {
        return func(mc *MiddlewareContext) *jsonrpc.Response {
            limiter := mc.Client.GetLimiter(mc.Method, cfg)

            if !limiter.Allow() {
                mc.Client.IncrementLimitViolation()

                if mc.Client.LimitViolations() >= cfg.BanThreshold {
                    mc.Client.Close("rate limit exceeded")
                }

                return jsonrpc.NewErrorResponse(mc.Request.ID,
                    errorsx.New(429, "TooManyRequests", "Rate limit exceeded"))
            }

            mc.Client.ResetLimitViolation()
            return next(mc)
        }
    }
}
```

### Metrics

```go
// pkg/ws/middleware/metrics.go

var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ws_requests_total",
            Help: "Total WebSocket requests",
        },
        []string{"method", "status"},
    )

    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "ws_request_duration_seconds",
            Help:    "WebSocket request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method"},
    )
)

func Metrics(next Handler) Handler {
    return func(mc *MiddlewareContext) *jsonrpc.Response {
        resp := next(mc)

        status := "success"
        if resp.Error != nil {
            status = "error"
        }

        requestsTotal.WithLabelValues(mc.Method, status).Inc()
        requestDuration.WithLabelValues(mc.Method).Observe(time.Since(mc.StartTime).Seconds())

        return resp
    }
}
```

## Hub 扩展

### 连接数限制

```go
// pkg/ws/hub.go

type HubConfig struct {
    MaxConnectionsPerIP   int
    MaxConnectionsPerUser int
}

func (h *Hub) CanAcceptConnection(ip, userID string) bool {
    if h.config.MaxConnectionsPerIP > 0 {
        if h.countByIP(ip) >= h.config.MaxConnectionsPerIP {
            return false
        }
    }
    if h.config.MaxConnectionsPerUser > 0 && userID != "" {
        if h.countByUser(userID) >= h.config.MaxConnectionsPerUser {
            return false
        }
    }
    return true
}
```

### 管理接口

```go
// pkg/ws/hub.go

// GetClient 通过 ID 获取客户端
func (h *Hub) GetClient(clientID string) *Client

// GetClientsByUser 获取用户的所有连接
func (h *Hub) GetClientsByUser(userID string) []*Client

// KickClient 踢掉指定连接
func (h *Hub) KickClient(clientID string, reason string) error

// KickUser 踢掉用户所有连接
func (h *Hub) KickUser(userID string, reason string) error
```

### 统计接口

```go
// pkg/ws/hub.go

type HubStats struct {
    TotalConnections   int64
    AuthenticatedConns int64
    AnonymousConns     int64
    ConnectionsByIP    map[string]int
    ConnectionsByUser  map[string]int
    ConnectionsByPlatform map[string]int
}

func (h *Hub) Stats() *HubStats
```

## 使用示例

### 路由注册

```go
// internal/apiserver/router/ws.go

func RegisterWSHandlers(router *ws.Router, biz biz.IBiz) {
    // 全局中间件
    router.Use(
        middleware.Recovery,
        middleware.RequestID,
        middleware.RateLimit(&middleware.RateLimitConfig{
            Default: 10,
            Methods: map[string]rate.Limit{
                "heartbeat": 0,  // 不限制
            },
            BanThreshold: 100,
        }),
        middleware.Metrics,
    )

    // 公开方法（无需认证）
    public := router.Group()
    public.Handle("heartbeat", handleHeartbeat)
    public.Handle("auth.login", authHandler.Login)

    // 需要认证的方法
    private := router.Group(middleware.Auth, middleware.Logger)
    private.Handle("subscribe", handleSubscribe)
    private.Handle("unsubscribe", handleUnsubscribe)
    private.Handle("user.info", userHandler.Info)
    private.Handle("user.update", userHandler.Update)
}
```

### Client 集成

```go
// pkg/ws/client.go

func (c *Client) handleMessage(data []byte) {
    var req jsonrpc.Request
    if err := json.Unmarshal(data, &req); err != nil {
        c.sendJSON(jsonrpc.NewErrorResponse(nil,
            errorsx.New(400, "ParseError", "Invalid JSON")))
        return
    }

    // 构建中间件上下文
    mc := &MiddlewareContext{
        Ctx:       c.ctx,
        Request:   &req,
        Client:    c,
        Method:    req.Method,
        StartTime: time.Now(),
    }

    // 通过路由器分发（自动执行中间件链）
    resp := c.router.Dispatch(mc)
    c.sendJSON(resp)
}
```

## 性能分析

| 组件 | 开销 |
|-----|------|
| 中间件链调用 | ~100ns |
| MiddlewareContext 分配 | ~50ns |
| Rate Limit 检查 | ~30ns |
| Metrics 记录 | ~200ns |
| **总计** | **~400ns** |

相比业务处理（1-10ms），中间件开销占比 < 0.1%，可忽略不计。

## 配置示例

```yaml
# configs/apiserver.yaml
websocket:
  enabled: true
  addr: ":8081"

  # 连接限制
  maxConnectionsPerIP: 10
  maxConnectionsPerUser: 5

  # 消息限流
  rateLimit:
    default: 10         # 默认每秒 10 条
    methods:
      heartbeat: 0      # 不限制
      subscribe: 5
      "user.*": 10
    banThreshold: 100   # 连续超限 100 次断开

  # Origin 检查
  allowedOrigins:
    - "https://example.com"
```

## 目录结构

```
pkg/ws/
├── client.go           # 客户端连接（已有，需修改）
├── hub.go              # 连接管理（已有，需扩展）
├── router.go           # 路由器（新增）
├── middleware.go       # 中间件类型定义（新增）
└── middleware/         # 内置中间件（新增）
    ├── recovery.go
    ├── requestid.go
    ├── logger.go
    ├── auth.go
    ├── ratelimit.go
    └── metrics.go
```

## 实施步骤

1. 定义 MiddlewareContext、Handler、Middleware 类型
2. 实现 Router 和 Group
3. 实现内置中间件（Recovery, RequestID, Logger, Auth, RateLimit, Metrics）
4. 扩展 Hub（连接限制、管理接口、统计接口）
5. 修改 Client.handleMessage() 使用 Router.Dispatch()
6. 迁移现有处理逻辑到新中间件架构
7. 添加配置支持
8. 编写测试

## 兼容性

- 对外 JSON-RPC 协议无变化
- 现有客户端无需修改
- 内部重构，渐进式迁移

## 实现状态

已完成：
- [x] 核心类型 (MiddlewareContext, Handler, Middleware, Chain)
- [x] Router 及 Group 分组支持
- [x] 内置中间件 (Recovery, RequestID, Auth, Logger, RateLimit)
- [x] Hub 管理 API (GetClient, GetClientsByUser, KickClient, KickUser, Stats)
- [x] Client 路由集成 (WithRouter 选项)
- [x] 内置 Handler (HeartbeatHandler, SubscribeHandler, UnsubscribeHandler)
- [x] apiserver 迁移到 Router 架构
- [x] 集成测试

待实现：
- [ ] Metrics 中间件（需要 Prometheus 依赖）
- [ ] 连接数限制 (MaxConnectionsPerIP, MaxConnectionsPerUser)
- [ ] BanThreshold 自动断开

关键文件：
- `pkg/ws/middleware.go` - 核心类型
- `pkg/ws/router.go` - Router 和 Group
- `pkg/ws/handlers.go` - 内置 Handler
- `pkg/ws/middleware/*.go` - 内置中间件
- `internal/apiserver/router/ws.go` - 路由注册
