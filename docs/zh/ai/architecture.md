# 架构与核心机制

本文档深入解析 AI 模块的内部架构、数据流转和核心可靠性机制。

## 1. 架构总览

AI 组件遵循 Bingo 的三层架构（Handler -> Biz -> Provider/Store），并引入了 **Registry 模式** 来解耦具体的模型厂商。

```
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│   Handler    │ ───► │     Biz      │ ───► │   Registry   │
│ (Controller) │      │ (Business)   │      │ (Providers)  │
└──────────────┘      └──────────────┘      └──────┬───────┘
       │                      │                    │
       ▼                      ▼                    ▼
 Validate Request     1. Load History       Select Provider
                      2. Reserve Quota    (OpenAI/Claude/...)
                      3. Load Agent
                      4. Save Session
```

## 2. 核心机制

### 2.1 上下文与会话 (Context & Session)

为了提供连续的对话体验，系统需要维护会话上下文。

- **自动加载**: 每次请求带上 `session_id`，Biz 层会自动从 MySQL 加载该会话的历史消息。
- **滑动窗口**: 为了防止 Token 超限和节省成本，系统实现了智能滑动窗口。
  - 配置 `Config.AI.Session.ContextWindow` 控制最大历史消息数。
  - **System Prompt 保护**: 在截断历史消息时，始终保留最开始的 System Prompt（如果存在），确保角色设定不丢失。
- **持久化**: 对话结束后，新的 User Message 和 Assistant Message 会异步写入数据库。

### 2.2 流式响应机制 (Streaming)

系统支持 Server-Sent Events (SSE) 标准，实现打字机效果。

- **协议**: `Content-Type: text/event-stream`
- **实现**:
  - `pkg/ai` 定义了 `ChatStream` 接口，返回一个只读 Channel。
  - Handler 层循环读取 Channel，将每个 Token 实时 flush 给客户端。
- **防泄露**: 监听 `ctx.Done()`，一旦客户端断开连接，立即取消上游 LLM 请求，释放 Goroutine 和配额资源。

### 2.3 高可用与重试 (Reliability & Retry)

针对 LLM API 常见的不稳定问题（如 503 Service Unavailable, Rate Limit），系统在 `pkg/ai/retry` 包中实现了自动重试机制。

- **策略**: 指数退避 (Exponential Backoff)。
- **触发条件**: 仅针对 **瞬时错误 (Transient Errors)** 重试，如网络超时、5xx 错误。对于 **永久错误**（如 401 Invalid Key, 400 Bad Request），直接失败，避免浪费资源。
- **范围**: 所有 Provider 的 `Chat` (普通对话) 接口均已内置重试。`ChatStream` 由于其实时性，通常由客户端控制重试，但服务端也会处理基础的连接错误。

### 2.4 配额系统 (Quota System)

基于 Redis 的高性能配额控制，保护系统不被滥用，并控制成本。

- **RPM (Requests Per Minute)**: 限制请求频率，防止刷接口。
- **TPD (Tokens Per Day)**: 限制每日 Token 消耗量，控制总预算。
- **自愈机制 (Self-Healing)**:
  采用 `Reserve` (预扣) -> `Use` (实耗) -> `Adjust` (调整) 模式。
  请求开始前先预扣估算的 Token，请求结束后根据真实消耗进行“多退少补”。配合 `defer` 机制，确保即使发生 Panic 或网络中断，预扣的配额也能最终被正确释放或修正。

## 3. 技术选型与决策 (Technology Decisions)

| 决策点 | 选择 | 理由 |
|--------|------|------|
| **底层框架** | **Eino** (CloudWeGo) | 字节跳动开源的 LLM 应用开发框架，Go 语言原生，类型安全，适合构建复杂的 AI 应用。 |
| **API 风格** | **OpenAI Compatible** | 行业事实标准，兼容性最强。前端可直接使用 OpenAI SDK，方便对接 LangChain 等生态。|
| **Provider 抽象** | **Interface 封装** | 位于 `pkg/ai`，不强依赖 Eino，保持核心业务逻辑的独立性，方便接入原生 SDK。 |
| **限流方案** | **Redis + GCRA** | 分布式一致性好，支持按用户动态调整配额 (RPM/TPD)。 |

