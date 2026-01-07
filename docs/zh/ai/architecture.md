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

## 2. 数据模型设计

### 2.1 Provider & Model 复合唯一键

`ai_model` 表使用 **`(provider_name, model)` 复合唯一键**，这是系统的核心设计决策之一。

**设计理由**：
- 不同 provider（如阿里云、火山引擎、硅基流动）可能托管相同的模型（如 `deepseek-r1`、`glm-4`）
- 同一模型名称在不同 provider 上可能有不同的定价和能力
- 支持灵活的 provider 切换和降级策略

**查询逻辑**：
```
用户请求 model="deepseek-r1"
    ↓
Store.FindActiveByModel("deepseek-r1")
    ↓ 返回多个匹配记录
按 sort ASC 排序 → 选择最高优先级的 provider
    ↓
Registry.Get(provider_name) → 获取 Provider 实例
```

**数据结构**：
```go
type AiModelM struct {
    ProviderName  string  // provider 名称（如 "openai", "qwen"）
    Model         string  // 模型名称（如 "gpt-4o", "deepseek-r1"）
    Status        string  // active / disabled
    Sort          int     // 优先级（数值越小越优先）
    AllowFallback bool    // 是否允许作为降级目标
    // ... 其他字段
}
// 复合唯一索引：uk_provider_model (provider_name, model)
```

### 2.2 Agent 智能体设计

智能体 (Agent) 是预定义的 AI 人格，包含独立的 System Prompt 和参数配置。

**核心特性**：
- `agent_id`: 全局唯一标识符，用于 API 调用
- `system_prompt`: 定义智能体行为的核心提示词
- `model`: 可选的强制绑定模型，优先级高于用户请求参数
- 参数覆盖逻辑: Agent 配置 > 用户请求 > 系统默认值

### 2.3 Session 会话管理

会话 (Session) 维护对话上下文，支持多轮对话的连续性。

**滑动窗口机制**：
- 配置 `Config.AI.Session.ContextWindow` 控制最大历史消息数
- **System Prompt 保护**: 截断历史时始终保留第一条 System Prompt
- 新消息异步持久化到数据库

## 3. 核心机制

### 3.1 上下文与会话 (Context & Session)

为了提供连续的对话体验，系统需要维护会话上下文。

- **自动加载**: 每次请求带上 `session_id`，Biz 层会自动从 MySQL 加载该会话的历史消息。
- **滑动窗口**: 为了防止 Token 超限和节省成本，系统实现了智能滑动窗口。
  - 配置 `Config.AI.Session.ContextWindow` 控制最大历史消息数。
  - **System Prompt 保护**: 在截断历史消息时，始终保留最开始的 System Prompt（如果存在），确保角色设定不丢失。
- **持久化**: 对话结束后，新的 User Message 和 Assistant Message 会异步写入数据库。

### 3.2 流式响应机制 (Streaming)

系统支持 Server-Sent Events (SSE) 标准，实现打字机效果。

- **协议**: `Content-Type: text/event-stream`
- **实现**:
  - `pkg/ai` 定义了 `ChatStream` 接口，返回一个只读 Channel。
  - Handler 层循环读取 Channel，将每个 Token 实时 flush 给客户端。
- **防泄露**: 监听 `ctx.Done()`，一旦客户端断开连接，立即取消上游 LLM 请求，释放 Goroutine 和配额资源。

### 3.3 高可用机制 (Reliability)

系统通过**自动重试**和**智能降级**两层机制保障服务可用性。

#### 3.3.1 自动重试 (Retry)

针对 LLM API 常见的不稳定问题（如 503 Service Unavailable, Rate Limit），系统在 `pkg/ai/retry` 包中实现了自动重试机制。

- **策略**: 指数退避 (Exponential Backoff)
- **触发条件**: 仅针对 **瞬时错误** 重试，如网络超时、5xx 错误
- **不触发**: 401/400 等永久错误直接失败，避免浪费资源
- **范围**: 所有 Provider 的 `Chat` 接口均已内置重试

#### 3.3.2 模型降级 (Fallback)

当请求的模型不可用或调用失败时，系统自动降级到备用模型，确保服务连续性。

**降级流程**：

```
ChatBiz.getProviderWithFallback()
    ↓
Store.FindActiveByModel(model)  → 返回 *model.AiModelM{ProviderName, Model}
    ↓ (失败)
FallbackSelector.SelectFallback()  → 按 sort 优先级选择备用模型
    ↓
Registry.Get(providerName)
    ↓ (失败且可重试)
最多再降级 1 次
```

**关键设计**：
- **复合唯一键**: `ai_model` 表使用 `(provider_name, model)` 复合唯一键，支持同一模型名称由多个 provider 提供
- **优先级控制**: 通过 `sort` 字段（数值越小优先级越高）控制 provider 选择顺序
- **降级限制**: 最多降级 1 次，避免级联失败
- **可配置性**: `allow_fallback` 字段控制模型是否允许作为降级目标

**触发降级的错误**：
- 模型未在 Registry 中注册（404）
- 429 (Rate Limit)
- 502/503/504 (网关错误)
- timeout / connection issues

**不触发降级的错误**：
- 401 (认证失败) - 配置错误，需人工处理
- quota 超限 - 需人工处理
- 参数错误 - 用户问题

#### 3.3.3 熔断器 (Circuit Breaker)

为防止故障 Provider 持续调用导致雪崩，系统为每个 Provider 实现了熔断器机制。

**三种状态**：

| 状态 | 说明 | 行为 |
|------|------|------|
| Closed | 正常状态 | 请求正常通过 |
| Open | 熔断状态 | 快速失败，触发降级 |
| Half-Open | 探测状态 | 允许少量请求探测恢复 |

**状态转换**：
```
Closed --[连续 5 次失败]--> Open
   ↓                           |
   <---[60 秒超时]-----------+

Open --[60 秒超时]--> Half-Open --[连续 2 次成功]--> Closed
                         ↓
                    [任何失败]--> Open
```

**配置**：
- `MaxFailures`: 5 - 触发熔断的连续失败次数
- `OpenTimeout`: 60s - 熔断持续时间
- `SuccessThreshold`: 2 - Half-Open 状态下恢复需要的连续成功次数

#### 3.3.4 健康检查 (Health Check)

系统后台定期对每个 Provider 进行健康检查，主动发现故障。

**检查机制**：
- **间隔**: 每 5 分钟
- **超时**: 30 秒
- **方式**: 发送最小测试请求（1 字符消息）
- **状态**: healthy / unhealthy / unknown

**查询接口**：
```go
chatBiz.HealthStatus() -> map[string]*ProviderHealth
```

---

### 3.4 配额系统 (Quota System)

基于 Redis 的高性能配额控制，保护系统不被滥用，并控制成本。

#### 3.4.1 RPM 限流

Redis 分钟级计数器，防止用户高频请求。

```go
key := "{app}:ai:rpm:{uid}:{minute_timestamp}"
count := Redis.Incr(key)
if count > rpm_limit {
    return ErrRateLimitExceeded
}
```

**特点**：
- 按分钟窗口计数，自动过期
- 超限时自动回滚计数
- 记录拒绝指标供监控

#### 3.4.2 TPD 配额

日级 Token 配额，支持用户级别和层级级别配置。

- 配额按天重置（UTC 零点或用户最后重置时间）
- Redis 缓存当天已用额度，减少 DB 查询
- 超限后返回 429 错误

#### 3.4.3 自愈机制

采用 `Reserve` (预扣) -> `Use` (实耗) -> `Adjust` (调整) 模式：

1. 请求开始前：预扣估算的 Token（默认 4096）
2. 请求结束后：根据真实消耗进行"多退少补"
3. 同步释放：配额未消耗时同步释放（5s 超时），确保可靠性

---

### 3.5 安全机制 (Security)

#### 3.5.1 输入验证

- **消息长度限制**: 单次请求总字符数不超过 15000（约 4000 tokens）
- **空消息检查**: 拒绝空消息请求
- **Session 验证**: 确保 Session 属于当前用户

#### 3.5.2 凭证管理

- API Key 支持环境变量配置（推荐生产环境）
- 配置文件使用注释提示环境变量用法
- 凭证不在日志中输出

---

### 3.6 可观测性 (Observability)

#### 3.6.1 Prometheus Metrics

系统暴露 8 个专用 AI Metrics，通过 `/metrics` 端点访问：

| Metric | 说明 |
|--------|------|
| `ai_request_duration_seconds` | 请求耗时（按 provider/model/stream 分组） |
| `ai_requests_total` | 请求总数（按状态分组：success/error） |
| `ai_fallback_total` | 降级激活次数（按 from/to provider 分组） |
| `ai_quota_reservation_total` | 配额操作计数（reserve/adjust/release） |
| `ai_circuit_breaker_state` | 熔断器状态（0=Open, 0.5=Half-Open, 1=Closed） |
| `ai_circuit_breaker_failures_total` | 熔断器拒绝次数 |
| `ai_rpm_rejections_total` | RPM 限流拒绝次数 |

#### 3.6.2 结构化日志

所有关键操作使用 `log.C(ctx).Infow/Errorw` 记录：
- 自动包含 TraceID（从 context 提取）
- 降级事件、熔断器状态变化、配额操作失败等

#### 3.6.3 健康状态 API

```go
chatBiz.HealthStatus() -> map[string]*ProviderHealth
```

返回每个 Provider 的健康状态：
```json
{
  "openai": {"status": "healthy", "last_check": "2026-01-07T10:00:00Z"},
  "claude": {"status": "unhealthy", "last_check": "2026-01-07T10:00:00Z", "last_error": "..."}
}
```

## 4. 技术选型与决策 (Technology Decisions)

| 决策点 | 选择 | 理由 |
|--------|------|------|
| **底层框架** | **Eino** (CloudWeGo) | 字节跳动开源的 LLM 应用开发框架，Go 语言原生，类型安全，适合构建复杂的 AI 应用。 |
| **API 风格** | **OpenAI Compatible** | 行业事实标准，兼容性最强。前端可直接使用 OpenAI SDK，方便对接 LangChain 等生态。|
| **Provider 抽象** | **Interface 封装** | 位于 `pkg/ai`，不强依赖 Eino，保持核心业务逻辑的独立性，方便接入原生 SDK。 |
| **限流方案** | **Redis + GCRA** | 分布式一致性好，支持按用户动态调整配额 (RPM/TPD)。 |

