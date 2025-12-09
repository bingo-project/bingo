# 整体架构设计

## 生态仓库

```
github.com/bingo-project/bingo/       # 框架核心
github.com/bingo-project/starter/     # 应用脚手架
github.com/bingo-project/websocket/   # 独立生态组件
github.com/bingo-project/bingoctl/    # CLI 工具
```

### component-base 处理

- cli 工具 → 迁移到 `bingoctl` 内部
- web 组件（jwt, openapi signer 等）→ 迁移到 `bingo/pkg`
- component-base 仓库废弃

### 生态拆分策略

| 包 | 归属 | 理由 |
|-----|------|------|
| `ws` | `bingo-project/websocket` | 独立价值高，可被其他框架使用 |
| `jsonrpc` | `websocket/pkg/jsonrpc` | 与 websocket 强相关 |
| `errorsx` | `bingo/pkg/errorsx` | 框架内部工具 |
| `jwt`, `signer` | `bingo/pkg/*` | 框架 web 组件 |

**拆分原则**：只有足够分量、有独立使用价值的包才值得单独建仓库。

## 框架分层架构

```
┌─────────────────────────────────────────────────────┐
│                    用户应用                          │
├─────────────────────────────────────────────────────┤
│  App（应用入口层）                                   │
│    └── 协调 Runnable 生命周期，统一启动入口          │
├─────────────────────────────────────────────────────┤
│  Runnable（可运行组件层）                            │
│    ├── HTTPServer                                   │
│    ├── GRPCServer                                   │
│    ├── WebSocketServer                              │
│    ├── TaskServer (asynq)                           │
│    └── 用户自定义 Runnable...                       │
├─────────────────────────────────────────────────────┤
│  pkg（通用工具层）                                   │
│    └── 框架内部工具，部分公开为 API                  │
└─────────────────────────────────────────────────────┘
```

## 核心概念

详细设计见 [concepts/](./concepts/) 目录：

| 概念 | 说明 | 详细设计 |
|------|------|----------|
| **Lifecycle** | 启动、运行、关闭的完整生命周期 | [lifecycle.md](./concepts/lifecycle.md) |
| **Runnable** | 可运行组件接口，借鉴 controller-runtime | [runnable.md](./concepts/runnable.md) |
| **App** | 框架统一入口，协调 Runnable 生命周期 | [app.md](./concepts/app.md) |
| **Health** | 健康检查，支持 /healthz 和 /readyz | [health.md](./concepts/health.md) |

## 配置驱动

### 框架配置 vs 应用配置

| 类型 | 内容 | 说明 |
|------|------|------|
| 框架配置 | App, Server, DB, Cache, Log, JWT | 框架提供，通用能力 |
| 应用配置 | Mail, Bot, Code, 业务配置... | 应用自定义 |

### 框架核心配置

```go
// 框架提供
type Config struct {
    App    AppConfig    `yaml:"app"`
    Server ServerConfig `yaml:"server"`
    DB     DBConfig     `yaml:"db"`
    Cache  CacheConfig  `yaml:"cache"`
    Log    LogConfig    `yaml:"log"`
    JWT    JWTConfig    `yaml:"jwt"`
}

type ServerConfig struct {
    HTTP      *HTTPConfig      `yaml:"http"`      // nil = 不启用
    GRPC      *GRPCConfig      `yaml:"grpc"`
    WebSocket *WebSocketConfig `yaml:"websocket"`
    Health    *HealthConfig    `yaml:"health"`    // 健康检查独立端口
}
```

### 应用扩展配置

```go
// 应用嵌入框架配置，添加业务配置
type MyAppConfig struct {
    bingo.Config `yaml:",inline"` // 嵌入框架配置
    Mail         MailConfig       `yaml:"mail"`
    Bot          BotConfig        `yaml:"bot"`
    // 其他业务配置...
}
```

### 配置加载

配置通过 `WithConfigName` 指定文件名（不含扩展名），默认为 `config`：

```go
// 单服务项目，加载 config.yaml
app, err := bingo.New()

// 多服务项目，加载 myapp-apiserver.yaml
app, err := bingo.New(
    bingo.WithConfigName("myapp-apiserver"),
)
```

**配置搜索顺序**（优先级从高到低）：

1. 命令行 `-c /path/to/config.yaml`
2. 环境变量 `{APPNAME}_CONFIG=/path/to/config.yaml`
3. 当前目录 `./config.yaml` 或 `./{configName}.yaml`
4. 标准路径 `/etc/{appname}/`、`~/.{appname}/`

**环境变量覆盖配置项**：

```bash
# config.yaml 中的 db.host 可被环境变量覆盖
MYAPP_DB_HOST=production-db
```

### 配置文件示例

```yaml
# 框架配置
app:
  name: myapp
  timezone: UTC

server:
  http:
    addr: ":8080"
  grpc:
    addr: ":9090"
  health:
    addr: ":8081"  # 健康检查独立端口
  # websocket 不配置 = 不启动

db:
  host: localhost:3306
  database: myapp

# 应用配置
mail:
  host: smtp.example.com
```

## 中间件策略

**核心逻辑复用，包装层各协议独立**。HTTP/gRPC/WebSocket 中间件签名本质不同，强行统一会导致抽象泄漏。

```go
// 核心逻辑（协议无关）
func ValidateToken(token string) (*Claims, error) { ... }

// HTTP 中间件（包装层）
func HTTPAuth() gin.HandlerFunc { ... }

// gRPC 中间件（包装层）
func GRPCAuth() grpc.UnaryServerInterceptor { ... }
```

### 内置中间件

框架提供常用中间件实现，参考 [go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)：

| 中间件 | 说明 |
|--------|------|
| Recovery | panic 恢复 |
| Logging | 请求日志 |
| Auth | 认证（用户提供 AuthFunc） |

## 链路追踪

### 内置（轻量）

框架默认启用轻量级追踪，无需配置：

| 字段 | 说明 |
|------|------|
| Request ID | 自动生成（UUID），或从 X-Request-ID header 获取 |
| Client IP | 从 X-Forwarded-For / X-Real-IP / 连接获取 |

```go
// 放入 context，各协议统一访问
requestID := bingo.RequestID(ctx)
clientIP := bingo.ClientIP(ctx)

// 日志自动带上
logger.Info("处理请求", "request_id", requestID, "client_ip", clientIP)
```

### 可选（完整 tracing）

复杂场景按需集成 OpenTelemetry：

```go
app.Add(otel.NewTracer(config))  // 按需启用
```

或使用 Service Mesh（Istio）在基础设施层处理。

## 测试辅助

框架不内置 mock 实现，但提供测试辅助函数：

```go
app := bingo.NewTestApp(
    bingo.WithDB(mockDB),
    bingo.WithLogger(mockLogger),
    bingo.WithConfig(testConfig),
)
```

### 等待就绪

测试时使用 `app.Ready()` channel 等待服务就绪：

```go
func TestAPI(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    app, _ := bingo.New()
    go app.Run(ctx)

    <-app.Ready()  // 等待就绪，不用 time.Sleep

    // 测试...
}
```

## 依赖处理策略

**原则**：通过接口抽象解耦，只公开必要的 API，框架提供默认实现。

| 包 | 处理方式 |
|-----|----------|
| `server` | 公开到 `pkg/server`，核心协议抽象 |
| `bootstrap` | 融入 App 层 |
| `log` | 复用现有 `internal/pkg/log.Logger` 接口，zap 实现 |
| `config` | 配置抽象，支持 viper |
| `auth` | 定义 `Authenticator` 接口 |
| `middleware` | HTTP/gRPC 中间件链 |
| `store`, `facade`, `task` | 保持 internal 或通过接口暴露必要部分 |

## 文件组织规范

- 文件夹统一用**单数**（符合 Go 惯例）
- 示例：`router/`, `handler/`, `server/`, `config/`

## 包组织

采用 controller-runtime 风格的包组织：**根包 re-export 常用类型，具体实现在子包**。

```
github.com/bingo-project/bingo/
├── bingo.go              # 根包，re-export 常用类型和函数
├── pkg/
│   ├── app/              # App、Runnable、Registrar 实现
│   ├── log/              # Logger 接口和 zap 实现
│   ├── server/           # HTTPServer、GRPCServer 等
│   ├── auth/             # 认证相关
│   ├── middleware/       # 中间件
│   └── signals/          # 信号处理
└── ...
```

**用户使用**：

```go
import "github.com/bingo-project/bingo"

// 大多数场景只需 import 根包
app, err := bingo.New(bingo.WithConfigName("myapp"))
app.Run(bingo.SetupSignalHandler())

// 需要细粒度控制时 import 子包
import "github.com/bingo-project/bingo/pkg/log"
log.SetLogger(myLogger)
```

## 框架与用户代码边界

| 归属 | 内容 |
|------|------|
| **框架提供** | App 层、Runnable/Registrar 接口、内置 Server、基础依赖、通用中间件、认证实现 |
| **用户负责** | 业务 Handler、路由注册、业务配置、业务中间件包装 |

框架提供通用实现（JWT auth、recovery 中间件等），用户可通过接口替换。
