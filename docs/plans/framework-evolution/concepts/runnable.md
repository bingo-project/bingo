# Runnable 设计

## 概述

Runnable 是框架的核心扩展接口，借鉴 controller-runtime 的设计。任何需要在 App 生命周期内运行的组件都实现此接口。

## 接口定义

```go
// Runnable：需要持续运行的组件（Server、Worker 等）
type Runnable interface {
    Start(ctx context.Context) error  // 阻塞运行，ctx 取消时退出
}

// Registrar：只需要启动前注册/配置的组件
type Registrar interface {
    Register(app *App) error
}

// Named：可选接口，用于日志输出
type Named interface {
    Name() string
}
```

### 设计理由

- **职责分离**：Runnable 用于运行，Registrar 用于注册，互不强制
- **单方法接口**：每个接口只有一个方法，简单直接
- **context 驱动**：Runnable 通过 `ctx.Done()` 触发关闭，Go 标准模式
- **Name() 可选**：框架检测是否实现 Named 接口，用于日志，不强制
- **借鉴 controller-runtime**：Go/K8s 生态惯例

### 两种组件的区别

| 类型 | 用途 | 生命周期 | 示例 |
|------|------|----------|------|
| Runnable | 需要持续运行 | 启动后阻塞，ctx 取消时退出 | HTTPServer, GRPCServer, TaskServer |
| Registrar | 启动前配置 | Register 执行完即结束 | RouteRegistrar, MiddlewareRegistrar |

组件可以同时实现两个接口（先注册，再运行）。

## 内置 Runnable

框架提供以下内置 Runnable：

| Runnable | 功能 | 说明 |
|----------|------|------|
| HTTPServer | HTTP 服务 | 支持 standalone 和 gateway 模式 |
| GRPCServer | gRPC 服务 | 标准 gRPC 服务 |
| WebSocketServer | WebSocket 服务 | 长连接服务 |
| TaskServer | 任务队列 | 基于 asynq，支持队列和定时任务 |

## 启动顺序

### 基本规则

1. **Register 阶段**：串行执行，建立依赖关系
2. **Start 阶段**：并发启动所有 Runnable

### 特殊情况：grpc-gateway 模式

HTTP Server 有两种模式：

```yaml
server:
  http:
    mode: standalone  # 纯 Gin，可并发启动
    # mode: gateway   # grpc-gateway，依赖 gRPC
```

**gateway 模式下**，Gateway Server 内部通过端口探测等待 gRPC 就绪：

```go
func (s *GatewayServer) Start(ctx context.Context) error {
    // 等待 gRPC 端口可连接
    if err := waitForPort(ctx, s.grpcAddr, 30*time.Second); err != nil {
        return fmt.Errorf("gRPC server not ready: %w", err)
    }

    // 继续启动 gateway...
}
```

**设计理由**：
- 端口探测简单可靠，10-100ms 延迟在生产环境可接受
- 保持 Runnable 接口最简（不需要 OnReady 回调机制）
- 依赖关系内聚在 Gateway 实现里，用户无需关心

框架会输出明确的日志：

```
INFO  starting gRPC server                     addr=:9090
INFO  starting HTTP server (gateway mode)      addr=:8080
INFO  waiting for gRPC server to be ready      addr=:9090
INFO  gRPC server ready, starting gateway
```

## 实现示例

### 简单 Runnable

```go
type MetricsServer struct {
    server *http.Server
}

func (s *MetricsServer) Start(ctx context.Context) error {
    errCh := make(chan error, 1)
    go func() {
        errCh <- s.server.ListenAndServe()
    }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        return s.server.Shutdown(shutdownCtx)
    }
}
```

### 纯 Registrar（只需要注册）

```go
type RouteRegistrar struct {
    routes []Route
}

// 只实现 Registrar 接口，无需实现 Runnable
func (r *RouteRegistrar) Register(app *App) error {
    engine := app.HTTPEngine()
    for _, route := range r.routes {
        engine.Handle(route.Method, route.Path, route.Handler)
    }
    return nil
}
```

### 同时实现两个接口（先注册，再运行）

```go
type TaskServer struct {
    server *asynq.Server
    mux    *asynq.ServeMux
}

// Register：注册任务处理器
func (s *TaskServer) Register(app *App) error {
    s.mux.HandleFunc("email:send", s.handleSendEmail)
    s.mux.HandleFunc("order:process", s.handleProcessOrder)
    return nil
}

// Start：运行任务服务器
func (s *TaskServer) Start(ctx context.Context) error {
    errCh := make(chan error, 1)
    go func() {
        errCh <- s.server.Run(s.mux)
    }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        s.server.Shutdown()
        return nil
    }
}
```

## 添加组件

```go
app := bingo.New()

// 添加 Runnable（需要运行的组件）
app.Add(&HTTPServer{...})
app.Add(&GRPCServer{...})

// 添加 Registrar（只需要注册的组件）
app.Register(&RouteRegistrar{...})
app.Register(&MiddlewareRegistrar{...})

// 同时实现两个接口的组件用 Add
// 框架会自动检测并在 Register 阶段调用 Register 方法
app.Add(&TaskServer{...})

app.Run(context.Background())
```

### Registrar 约定

`Register()` **不应有外部副作用**，只做内存注册：

```go
// ✅ 正确：只注册路由
func (r *RouteRegistrar) Register(app *App) error {
    app.HTTPEngine().POST("/orders", r.handler)
    return nil
}

// ❌ 错误：在 Register 里做外部操作
func (r *BadRegistrar) Register(app *App) error {
    createDatabaseTable()  // 不应该在这里
    return nil
}
```

外部操作（创建表、写文件等）应在 `Start()` 里或单独的初始化脚本中。

## 决策记录

| 主题 | 决策 | 理由 |
|------|------|------|
| 接口设计 | Runnable 和 Registrar 独立，互不强制 | 职责分离，避免空实现 |
| Runnable 接口 | 只有 Start(ctx)，Name() 作为可选 Named 接口 | 保持最简 |
| 添加方式 | Add(Runnable) 和 Register(Registrar) 两个方法 | 类型安全，API 意图明确 |
| 启动顺序 | Register 串行 → Start 并发 | Register 有依赖，Server 间一般无依赖 |
| Register 执行顺序 | 按添加顺序 | 直观可预测 |
| grpc-gateway 依赖 | 端口探测，Gateway 内部等待 | 简单可靠，接口最简 |
| 关闭方式 | ctx.Done() 触发 | 统一机制，Go/K8s 惯例 |
| Register 约定 | 不应有外部副作用 | 只做内存注册 |
| 并发安全 | Add/Register 不支持并发调用 | 实际场景不需要 |
