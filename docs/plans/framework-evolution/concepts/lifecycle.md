# 生命周期管理

## 概述

框架采用 context 驱动的生命周期管理，借鉴 controller-runtime 的设计。context 取消即触发关闭流程。

## 完整生命周期

```
┌─────────────────────────────────────────────────────────────────┐
│                         启动阶段                                 │
├─────────────────────────────────────────────────────────────────┤
│  1. 健康检查端点立即可用（/healthz 返回 200）                     │
│  2. Registrar.Register()  ──  按添加顺序串行执行                 │
│  3. Runnable.Start()      ──  并发启动所有组件                   │
│  4. 标记就绪（/readyz 返回 200，close Ready channel）            │
├─────────────────────────────────────────────────────────────────┤
│                         运行阶段                                 │
├─────────────────────────────────────────────────────────────────┤
│  处理请求，健康检查持续运行                                      │
├─────────────────────────────────────────────────────────────────┤
│                         关闭阶段                                 │
├─────────────────────────────────────────────────────────────────┤
│  1. context 取消（用户通过 SetupSignalHandler 或手动取消）        │
│  2. /readyz 返回 503（从负载均衡摘除）                           │
│  3. 等待所有 Runnable 退出                                       │
│  4. App 层超时兜底，超时后强制返回                                │
└─────────────────────────────────────────────────────────────────┘
```

## 信号处理

**框架不自动处理信号**，由用户控制，类似 controller-runtime：

```go
// 标准用法
app.Run(bingo.SetupSignalHandler())

// 自定义场景
ctx, cancel := context.WithCancel(context.Background())
// 用户自己处理信号...
app.Run(ctx)
```

`SetupSignalHandler()` 的行为：
1. 注册 SIGTERM 和 SIGINT 监听
2. 返回 context，收到第一个信号时 cancel
3. 收到第二个信号时直接 `os.Exit(1)`（强制退出）

```go
// pkg/signals/signals.go
func SetupSignalHandler() context.Context {
    ctx, cancel := context.WithCancel(context.Background())

    c := make(chan os.Signal, 2)
    signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

    go func() {
        <-c
        cancel()      // 第一次信号：优雅关闭
        <-c
        os.Exit(1)    // 第二次信号：强制退出
    }()

    return ctx
}
```

## Context 驱动

所有生命周期通过 context 控制：

```go
func (app *App) Run(ctx context.Context) error {
    // 1. Register 阶段（按添加顺序串行）
    for _, component := range app.components {
        if reg, ok := component.(Registrar); ok {
            if err := reg.Register(app); err != nil {
                return err
            }
        }
    }

    // 2. 启动所有 Runnable（并发）
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    var g errgroup.Group
    for _, component := range app.components {
        if r, ok := component.(Runnable); ok {
            g.Go(func() error { return r.Start(ctx) })
        }
    }

    // 3. 标记就绪
    app.markReady()

    // 4. 等待退出或超时
    done := make(chan error, 1)
    go func() { done <- g.Wait() }()

    select {
    case err := <-done:
        return err
    case <-time.After(app.shutdownTimeout):
        return errors.New("shutdown timeout")
    }
}
```

## Runnable 实现关闭

每个 Runnable 通过监听 `ctx.Done()` 触发关闭。**Runnable 不设置自己的超时**，由 App 层统一兜底：

```go
func (s *HTTPServer) Start(ctx context.Context) error {
    // 启动监听
    errCh := make(chan error, 1)
    go func() {
        errCh <- s.server.ListenAndServe()
    }()

    // 等待关闭信号或错误
    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        // 优雅关闭，不设超时，由 App 层兜底
        return s.server.Shutdown(context.Background())
    }
}
```

## Shutdown 超时

**App 层统一超时兜底**，适配 K8s `terminationGracePeriodSeconds`：

```go
app := bingo.New(
    bingo.WithShutdownTimeout(30 * time.Second),  // 默认 30s
)
```

**设计理由**：
- 职责清晰：Runnable 负责优雅关闭，App 负责超时兜底
- 与 K8s 对齐：K8s 的 `terminationGracePeriodSeconds` 是进程级超时
- Runnable 实现更简单，不需要关心超时配置

## 错误处理

### 启动失败

任一 Runnable 启动失败，取消 context，其他 Runnable 收到信号后退出：

```go
var g errgroup.Group
ctx, cancel := context.WithCancel(ctx)
defer cancel()

for _, r := range app.runnables {
    g.Go(func() error {
        if err := r.Start(ctx); err != nil {
            cancel()  // 通知其他 Runnable 退出
            return err
        }
        return nil
    })
}
```

### 关闭失败

继续关闭其他组件，返回第一个错误：

```go
// errgroup 返回第一个非 nil 错误
// 即使某个 Runnable 关闭失败，其他的仍会继续关闭
```

## 决策记录

| 主题 | 决策 | 理由 |
|------|------|------|
| 生命周期控制 | context 驱动 | Go/K8s 生态惯例 |
| 信号处理 | 用户控制，提供 SetupSignalHandler() | 类似 controller-runtime，用户有完全控制权 |
| 启动顺序 | Register 串行 → Start 并发 | Register 有依赖，Server 间无依赖 |
| Register 执行顺序 | 按添加顺序 | 直观可预测，用户代码顺序即执行顺序 |
| 关闭触发 | ctx.Done() | 统一机制，无需显式 Shutdown 方法 |
| Shutdown 超时 | App 层统一兜底 | Runnable 不设超时，App 做最终超时控制 |
| 启动失败 | 取消 context，其他退出 | fail fast |
| 关闭失败 | 继续关闭其他，返回首个错误 | best-effort |
