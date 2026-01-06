# AI Model Fallback Design

## Overview

AI 模型降级机制：当请求的模型不可用或调用失败时，自动降级到备用模型，确保服务可用性。

## Design Decisions

### 1. 降级策略

**选择：混合方案（可配置降级行为）**

- 在 `ai_model` 表新增 `allow_fallback` 字段
- 每个模型独立配置是否允许作为降级目标
- 默认值 `true`，降级默认启用
- 特殊模型（如企业专有模型）可手动关闭

### 2. 降级触发条件

**选择：模型未注册 + Provider 特定错误**

触发降级的错误：
- 模型未在 Registry 中注册（404）
- 429 (Rate Limit)
- 502/503/504 (网关错误)
- timeout / connection issues

不触发降级的错误：
- 401 (认证失败) - 配置错误
- quota 超限 - 需要人工处理
- 参数错误 - 用户问题

### 3. 降级次数上限

**最多 1 次降级**，避免级联失败到完全不合适的模型。

### 4. 模型优先级

**使用现有 `sort` 字段**，不需要新增字段。数据库已按 `sort ASC, id ASC` 排序。

### 5. 降级行为记录

**仅日志记录**，不暴露给用户。使用结构化日志记录降级事件。

### 6. 失败返回

**用户友好错误**：返回 `ErrAIAllModelsFailed`，错误消息"服务暂时不可用，请稍后重试"。

## Architecture

```
ChatBiz → Registry.GetByModel (优先)
          ↓ (失败)
       FallbackSelector.SelectFallback (按 sort 顺序)
          ↓
       Registry.GetByModel (降级模型)
          ↓
       Provider.Chat/ChatStream
          ↓ (失败且可重试)
       最多再降级 1 次
```

## Implementation

### 组件位置

```
internal/pkg/ai/
├── fallback.go          # 新增：降级选择逻辑
├── loader.go            # 现有
├── ai.go                # 修改：导出 FallbackSelector
└── ...
```

### 核心接口

```go
type FallbackSelector struct {
    store    store.AiModelStore
    registry *Registry
}

func (s *FallbackSelector) SelectFallback(ctx context.Context, originalModel string) string
```

### 降级选择逻辑

```
1. 从 ai_model 表获取所有 active 状态的模型（按 sort ASC）
2. 过滤：排除原始模型
3. 过滤：只保留 allow_fallback=true 的模型
4. 检查：模型在 Registry 中已注册
5. 返回：第一个满足条件的模型
```

## Database Changes

### Migration

修改 `internal/pkg/database/migration/2025_12_29_100001_create_ai_model_table.go`：

```go
type CreateAIModelTable struct {
    // ... 现有字段
    AllowFallback bool `gorm:"type:tinyint(1);not null;default:1;comment:是否允许作为降级目标"`
    // ...
}
```

### Model

```go
// internal/pkg/model/ai_model.go

type AiModelM struct {
    // ...
    AllowFallback bool `gorm:"column:allow_fallback;..." json:"allowFallback"`
    // ...
}
```

## Error Handling

### 新增错误码

```go
// internal/pkg/errno/ai.go

var ErrAIAllModelsFailed = &errorsx.ErrorX{
    Code:    503,
    Reason:  "ServiceUnavailable.AllModelsFailed",
    Message: "AI service is temporarily unavailable, please try again later.",
}
```

### retry.go 修复

移除 `insufficient_quota` 从可重试错误列表（配额不足不应重试）。

## File Changes Summary

| 类型 | 文件 | 变更 |
|------|------|------|
| 新增 | `internal/pkg/ai/fallback.go` | 降级选择逻辑 |
| 修改 | `internal/pkg/ai/ai.go` | 导出 FallbackSelector |
| 修改 | `internal/apiserver/biz/chat/chat.go` | 注入 FallbackSelector，改造 Chat/ChatStream |
| 修改 | `internal/pkg/model/ai_model.go` | 新增 AllowFallback 字段 |
| 修改 | `internal/pkg/database/migration/..._create_ai_model_table.go` | 同步字段 |
| 修改 | `pkg/ai/retry.go` | 移除 insufficient_quota 重试 |
| 修改 | `internal/pkg/errno/ai.go` | 新增 ErrAIAllModelsFailed |

## Testing

| 层级 | 方式 | 覆盖场景 |
|------|------|----------|
| fallback_test.go | Mock Store + Registry | 正常降级、无可用降级、allow_fallback=false |
| chat_test.go | Mock Store + Registry | GetByModel 失败降级、Provider 调用失败降级 |
| E2E | 真实环境 | 完整降级流程 |

## Verification Steps

```bash
# 1. 数据库迁移
bingo migrate reset
bingo migrate up
bingo db seed

# 2. 构建
make build

# 3. 测试
make test ./internal/pkg/ai/...
make test ./internal/apiserver/biz/chat/...

# 4. 代码检查
make lint
```
