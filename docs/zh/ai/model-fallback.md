# AI Model Fallback

AI 模型降级机制：当请求的模型不可用或调用失败时，自动降级到备用模型，确保服务可用性。

## 架构

```
ChatBiz → Store.FindActiveByModel → Registry.Get (优先)
          ↓ (失败)
       FallbackSelector.SelectFallback (按 sort 顺序)
          ↓
       Registry.Get (降级模型的 provider)
          ↓
       Provider.Chat/ChatStream
          ↓ (失败且可重试)
       最多再降级 1 次
```

## 降级策略

| 配置 | 说明 |
|------|------|
| `allow_fallback` | 模型在 `ai_model` 表的字段，控制是否允许作为降级目标（默认 `true`） |
| `sort` | 模型优先级，数值越小优先级越高 |
| 最大降级次数 | 1 次，避免级联失败 |

## 触发条件

**触发降级的错误：**
- 模型未在 Registry 中注册（404）
- 429 (Rate Limit)
- 502/503/504 (网关错误)
- timeout / connection issues

**不触发降级的错误：**
- 401 (认证失败) - 配置错误，需人工处理
- quota 超限 - 需人工处理
- 参数错误 - 用户问题

## 组件

### FallbackSelector

位置：`internal/pkg/ai/fallback.go`

**降级选择逻辑：**
1. 从 `ai_model` 表获取所有 `active` 状态的模型（按 `sort ASC`）
2. 过滤：排除原始模型
3. 过滤：只保留 `allow_fallback=true` 的模型
4. 检查：模型在 Registry 中已注册
5. 返回：第一个满足条件的模型

## 错误处理

所有模型失败时返回 `ErrAIAllModelsFailed`，HTTP 503 状态码，提示用户"服务暂时不可用，请稍后重试"。

## 日志

降级事件会记录结构化日志：

```
INFO AI model fallback selected original=gpt-4o fallback=gpt-3.5-turbo
INFO AI provider error, using fallback model=gpt-4o fallback=gpt-3.5-turbo err=...
INFO AI provider stream error, using fallback model=gpt-4o fallback=gpt-3.5-turbo err=...
```
