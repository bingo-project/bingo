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
│              (Proto message 作为参数/返回值)                  │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Store 层                                │
└─────────────────────────────────────────────────────────────┘
```

## 目录结构

```
internal/apiserver/
├── biz/                    # 业务逻辑（协议无关）
│   ├── biz.go              # interface 定义
│   └── user/
│       └── user.go         # 实现，参数用 proto message
│
├── handler/                # 协议处理器（可插拔）
│   ├── http/               # 独立 HTTP（Gin）
│   ├── grpc/               # gRPC Handler
│   ├── gateway/            # gRPC-Gateway（可选）
│   └── ws/                 # WebSocket (JSON-RPC 2.0)
│
├── server/                 # 服务组装
│   └── server.go           # 根据配置组装协议
│
└── store/                  # 数据访问

pkg/proto/                  # Proto 定义（数据层）
├── user/v1/
│   ├── user.proto          # 消息定义（所有模式共用）
│   └── user_service.proto  # 服务定义（gRPC/Gateway 用）
└── common/
    └── error.proto         # 统一错误类型（可选）
```

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

```go
// internal/apiserver/server/server.go
func Run(cfg *config.Config, biz biz.IBiz) error {
    var servers []Server

    // gRPC（如果启用）
    if cfg.Server.GRPC.Enabled {
        servers = append(servers, grpc.NewServer(cfg, biz))
    }

    // HTTP
    if cfg.Server.HTTP.Enabled {
        switch cfg.Server.HTTP.Mode {
        case "gateway":
            servers = append(servers, gateway.NewServer(cfg))
        default:
            servers = append(servers, http.NewServer(cfg, biz))
        }
    }

    // WebSocket
    if cfg.Server.WebSocket.Enabled {
        servers = append(servers, ws.NewServer(cfg, biz))
    }

    return runAll(servers)
}
```

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
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    resp, err := h.biz.User().Login(c.Request.Context(), &req)
    core.WriteResponse(c, err, resp)
}
```

### WebSocket Handler

WebSocket 采用 JSON-RPC 2.0 规范，详见 [WebSocket 统一方案](websocket-unification.md)。

```go
func RegisterUserHandlers(a *adapter.Adapter, biz biz.IBiz) {
    adapter.Register(a, "user.login", biz.User().Login, &pb.LoginRequest{})
    adapter.Register(a, "user.get", biz.User().Get, &pb.GetUserRequest{})
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

- [WebSocket 统一方案](websocket-unification.md) - JSON-RPC 2.0 消息格式、适配器实现
- [gRPC-Gateway 集成](grpc-gateway.md) - Gateway 模式配置与使用
- [统一错误处理](unified-error-handling.md) - 三协议错误格式统一
- [认证中间件迁移](auth-middleware-migration.md) - 统一认证实现
