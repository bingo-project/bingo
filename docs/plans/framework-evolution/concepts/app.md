# App 层设计

## 概述

App 是框架的统一入口，借鉴 controller-runtime 的 Manager 模式，协调 Runnable 生命周期并提供依赖访问。

## 创建与启动

### New() - 创建 App 并加载配置

```go
// 创建 App，解析配置文件
// 返回 error，不会 os.Exit（可测试）
func New(opts ...Option) (*App, error)
```

`New()` 职责：
- 解析配置文件
- 校验配置格式
- **不**初始化 DB、Cache 等依赖（那是 Init 的职责）

```go
// 单服务项目，加载 config.yaml
app, err := bingo.New()

// 多服务项目，显式指定配置文件名
app, err := bingo.New(
    bingo.WithConfigName("myapp-apiserver"),
)
```

### Init() - 初始化基础依赖

```go
// 初始化依赖（Logger, DB, Cache 等），不启动 Server
func (app *App) Init() error
```

`Init()` 职责：
- 初始化基础依赖（Config, Logger, DB, Cache）
- **不**创建 Server（那是 Run 的职责）
- CLI 场景轻量，不创建不需要的组件

### Run() - 启动服务

```go
// 阻塞启动（Init + 启动 Runnable + 等待退出 + Close）
func (app *App) Run(ctx context.Context) error
```

`Run()` 职责：
- 自动调用 `Init()`（如果还没调用过）
- 执行 Register 阶段
- 启动所有 Runnable
- 标记就绪
- 等待 ctx 取消或 Runnable 退出
- 清理资源

### Close() - 清理资源

```go
// 清理资源（关闭 DB 连接等）
func (app *App) Close() error
```

### Ready() - 就绪通知

```go
// 返回 channel，就绪后 close
func (app *App) Ready() <-chan struct{}
```

## 使用示例

### 服务场景

```go
func main() {
    app, err := bingo.New(
        bingo.WithConfigName("myapp-apiserver"),
    )
    if err != nil {
        log.Fatal(err)
    }

    if err := app.Run(bingo.SetupSignalHandler()); err != nil {
        log.Fatal(err)
    }
}
```

### CLI 场景

```go
func main() {
    app, err := bingo.New(
        bingo.WithConfigName("myapp-apiserver"),
    )
    if err != nil {
        log.Fatal(err)
    }

    if err := app.Init(); err != nil {
        log.Fatal(err)
    }
    defer app.Close()

    // 使用依赖执行任务
    if err := migration.Run(app.DB()); err != nil {
        log.Fatal(err)
    }
}
```

### 测试场景

```go
func TestAPI(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    app, _ := bingo.New()
    go app.Run(ctx)

    <-app.Ready()  // 等待就绪，不用 time.Sleep

    resp, _ := http.Get("http://localhost:8080/api")
    // 断言...

    cancel()  // 触发关闭
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

// 同时实现 Runnable + Registrar 的组件用 Add
app.Add(&TaskServer{...})

// 链式调用
app.Add(server1).Add(server2).Register(routes).Run(ctx)
```

### API 签名

```go
func (app *App) Add(r Runnable) *App       // 需要运行的组件
func (app *App) Register(r Registrar) *App // 只需要注册的组件
```

## 依赖访问

### 设计原则

参考 controller-runtime 的 Manager 模式：

1. **类型化方法优于通用容器**：`app.DB()` 优于 `app.Get("db")`
2. **不做运行时容器**：Go 生态用 Wire 做编译期注入更合适
3. **显式传递优于隐式注入**：构造函数传入依赖

### 类型化访问

**统一使用方法访问**，不暴露公开字段：

```go
type App struct {
    config *Config    // 私有
    logger Logger     // 私有
    db     *gorm.DB   // 私有
    cache  Cache      // 私有
    // ...
}

// 所有依赖通过方法访问
func (app *App) Config() *Config           // 必需，总是存在
func (app *App) Logger() Logger            // 必需，总是存在
func (app *App) DB() *gorm.DB              // 可选，未配置时 panic
func (app *App) Cache() Cache              // 可选，未配置时 panic
func (app *App) HTTPEngine() *gin.Engine   // 可选，未配置时 panic
func (app *App) GRPCServer() *grpc.Server  // 可选，未配置时 panic
```

### 为什么统一用方法而非字段

| 考量 | 方法 | 字段 |
|------|------|------|
| 可加逻辑 | ✅ panic 检查、懒加载 | ❌ 无法加逻辑 |
| 封装性 | ✅ 私有字段，不可被修改 | ❌ 公开字段可被外部修改 |
| API 一致性 | ✅ 统一风格 | ❌ 需要记哪些是字段 |
| 未来扩展 | ✅ 可改实现不影响调用方 | ❌ 字段类型固定 |

### 可选组件的访问策略

借鉴 controller-runtime：**构造时校验，访问时信任**。

可选组件（DB, Cache 等）未配置时调用会 panic：

```go
func (app *App) DB() *gorm.DB {
    if app.db == nil {
        panic("database not configured")
    }
    return app.db
}
```

**理由**：
- 如果代码调用 `app.DB()`，说明业务依赖数据库
- 没配置是**配置错误**，应该 fail-fast
- 避免到处 nil check，代码更简洁

### 在 Registrar 中使用依赖

```go
type OrderHandler struct {
    db     *gorm.DB
    logger Logger
}

func (h *OrderHandler) Register(app *App) error {
    h.db = app.DB()
    h.logger = app.Logger()

    app.HTTPEngine().POST("/orders", h.Create)
    return nil
}
```

## 配置选项

```go
app := bingo.New(
    bingo.WithConfig(cfg),
    bingo.WithLogger(logger),
    bingo.WithShutdownTimeout(30 * time.Second),
)
```

## 测试辅助

```go
app := bingo.NewTestApp(
    bingo.WithDB(mockDB),
    bingo.WithLogger(mockLogger),
    bingo.WithConfig(testConfig),
)

handler := order.NewHandler(app.DB(), app.Logger())
// 测试 handler...
```

## 决策记录

| 主题 | 决策 | 理由 |
|------|------|------|
| New() 职责 | 解析配置，返回 error | 可测试，库不应直接 os.Exit |
| Init() 职责 | 初始化基础依赖（DB/Cache） | Server 在 Run 时创建，CLI 场景轻量 |
| Run() 行为 | 阻塞，ctx 取消触发关闭 | context 驱动，Go 风格 |
| CLI 支持 | Init() + Close()，不强制 Run() | CLI 工具也能复用依赖管理 |
| 就绪通知 | Ready() 返回 channel | 测试可精确等待，不用 time.Sleep |
| 关闭方式 | 无显式 Shutdown，通过 ctx 取消 | 简化 API |
| 依赖访问 | 统一方法访问，不暴露字段 | API 一致，可加逻辑，封装性好 |
| 可选组件 | 未配置时 panic | Fail-fast，stack trace 足够定位 |
| 配置文件 | WithConfigName 指定，默认 "config" | 多服务需显式指定 |
| 设计参考 | controller-runtime Manager | Go 生态惯例 |
| Shutdown 超时 | WithShutdownTimeout 配置 | App 统一兜底 |
