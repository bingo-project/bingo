# AI Chat 组件

AI Chat 是 Bingo 框架提供的核心 AI 对话能力组件，支持多模型厂商（OpenAI, Claude, Gemini, Qwen 等）、流式响应、自动上下文管理及配额控制。

## 架构设计

本组件严格遵循 Bingo 的三层架构（Handler -> Biz -> Provider/Store），并在此基础上引入了 **Registry 模式** 来管理异构的 AI 供应商。

### 核心流程图

```
┌──────────────────────────────────────────────────────────┐
│                        User (Client)                     │
└──────────────────────────────────────────────────────────┘
             │ HTTP / WebSocket (Stream)
             ▼
┌──────────────────────────────────────────────────────────┐
│                internal/apiserver/handler                │
│             (Request Parsing, Validation)                │
└──────────────────────────────────────────────────────────┘
             │ *ai.ChatRequest
             ▼
┌──────────────────────────────────────────────────────────┐
│                 internal/apiserver/biz                   │
│                                                          │
│  ┌─────────────────┐   1. Load History  ┌─────────────┐  │
│  │ Context Manager │ ◄───────────────── │  MySQL DB   │  │
│  │ (Sliding Window)│                    └─────────────┘  │
│  └─────────────────┘                                     │
│          │                                               │
│          ▼ 2. Reserve Quota                              │
│  ┌─────────────────┐                    ┌─────────────┐  │
│  │   Quota System  │ ◄── Grant/Deny ──► │    Redis    │  │
│  │   (RPM / TPD)   │                    └─────────────┘  │
│  └─────────────────┘                                     │
│          │                                               │
│          ▼ 3. Select Provider                            │
│  ┌─────────────────┐                                     │
│  │   AI Registry   │                                     │
│  └─────────────────┘                                     │
│          │                                               │
└──────────┼───────────────────────────────────────────────┘
           │ 4. Call Provider
           ▼
┌──────────────────────────────────────────────────────────┐
│                        pkg/ai                            │
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  OpenAI  │  │  Claude  │  │  Gemini  │  │   Qwen   │  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
└──────────────────────────────────────────────────────────┘
           │
           ▼
      Cloud AI APIs
```

## 核心机制

### 1. 多厂商管理 (Provider Registry)

为了屏蔽不同 AI 厂商 API 的差异，我们定义了统一的 `Provider` 接口，并通过 `Registry` 进行管理。

- **接口定义** (`pkg/ai/provider.go`): 统一了 `Chat`, `ChatStream`, `Name`, `Models` 等方法。
- **注册机制**: 在应用启动时 (`internal/apiserver/app.go`)，将配置中启用的 Provider 注册到全局 Registry。
- **动态路由**: 请求可指定 Model，Registry 会根据 Model ID 自动路由到对应的 Provider。

### 2. 上下文与会话 (Context & Session)

组件自动处理多轮对话的上下文维护：

- **自动加载**: 根据 `session_id` 自动从数据库加载历史消息。
- **滑动窗口**: 基于 `Config.AI.Session.ContextWindow` 配置，自动截断过早的消息，但保留 System Prompt。
- **消息去重**: 智能检测前端重发的历史消息，防止数据库存储重复数据。

### 3. 配额系统 (Quota System)

基于 Redis 的高性能配额控制，支持两种维度：

- **RPM (Requests Per Minute)**: 限制请求频率。
- **TPD (Tokens Per Day)**: 限制每日 Token 消耗量。

**自愈机制 (Self-Healing)**:
采用 `Reserve` -> `Use` -> `Adjust` 模式。先预扣配额，请求结束后根据实际消耗进行“多退少补”。配合 `defer` 机制，确保即使在 Panic 或网络中断时也能正确释放预扣配额。

## 配置说明

在 `configs/bingo-apiserver.yaml` 中配置：

```yaml
ai:
  default_model: "gpt-4o"
  proxy_url: "http://127.0.0.1:7890" # 可选代理
  
  session:
    max_messages: 50      # 单次加载最大历史数
    context_window: 10    # 发送给 AI 的最大历史数
    
  providers:
    openai:
      enabled: true
      api_key: "sk-..."
      models:
        - id: "gpt-4o"
          usage_price: 1.0
          
    claude:
      enabled: true
      api_key: "sk-ant-..."
      
    gemini:
      enabled: true
      api_key: "AIza..."
```

## 开发指南

### 添加新的 AI Provider

1.  **实现接口**: 在 `pkg/ai/providers/<name>/` 下创建新包，实现 `ai.Provider` 接口。
2.  **复用工具**: 使用 `pkg/ai/common.go` 中的 `ConvertMessages`, `GenerateID` 等工具函数减少重复代码。
3.  **注册**: 在 `internal/apiserver/http.go` 的 `initAIRegistry` 中添加初始化逻辑。
4.  **配置**: 在 `internal/pkg/config` 和配置文件中添加对应配置项。

### 调试

- **日志**: 所有关键步骤均有 `log.C(ctx)` 结构化日志。
- **Token 监控**: 搜索日志关键字 `AdjustTPD` 可查看每次请求的实际 Token 消耗与配额调整情况。
