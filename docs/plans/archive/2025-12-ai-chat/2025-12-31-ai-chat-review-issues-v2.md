# AI Chat 实现 Review 问题清单 v2

> 2025-12-31 代码审查发现的问题

## Critical (必须修复)

- [ ] **1. Provider 实现有大量重复代码** - `pkg/ai/providers/`
  - `openai/provider.go`、`claude/provider.go`、`gemini/provider.go`、`qwen/provider.go` 中的 `convertMessages`、`convertResponse`、`convertStreamChunk`、`generateID`、`extractUsage` 函数完全相同
  - **影响**: 维护成本高，修改一处需要同步四处
  - **修复**: 将公共函数提取到 `pkg/ai/provider_common.go` 或创建 base provider

- [ ] **2. `saveStreamToSession` 保存用户消息逻辑有误** - `internal/apiserver/biz/chat/chat.go:237-250`
  ```go
  // 遍历 req.Messages 保存用户消息
  for _, msg := range req.Messages {
      if msg.Role == ai.RoleUser {
          // 保存...
      }
  }
  ```
  - **问题**: `req.Messages` 在 `loadAndMergeHistory` 后已经包含了历史消息，会重复保存
  - **影响**: 同一条用户消息会被多次保存到数据库
  - **修复**: 应只保存新消息（调用前传入的原始消息）

- [ ] **3. `saveToSession` 同样的问题** - `internal/apiserver/biz/chat/chat.go:383-396`
  - 遍历 `req.Messages` 保存用户消息，但此时已包含历史消息
  - **影响**: 同上，重复保存
  - **修复**: 同上

## Important (应该修复)

- [ ] **4. Gemini Provider New 函数签名不一致** - `pkg/ai/providers/gemini/provider.go:30`
  ```go
  func New(ctx context.Context, cfg *Config) (*Provider, error)
  ```
  - 其他 Provider 都是 `func New(cfg *Config) (*Provider, error)`
  - **影响**: 初始化代码需要特殊处理 Gemini
  - **修复**: 统一签名，内部在需要时使用 `context.Background()`

- [ ] **5. `loadAndMergeHistory` 可能有重复消息** - `internal/apiserver/biz/chat/chat.go:321-376`
  - 如果前端同时传 `session_id` 和历史消息，会与数据库加载的历史消息合并
  - **影响**: 消息列表会越来越长
  - **修复**: 当有 `session_id` 时，忽略请求中的历史消息，只使用新的最后一条用户消息

- [ ] **6. 流式响应 quota 调整的竞态条件** - `internal/apiserver/biz/chat/chat.go:206-212`
  ```go
  // Adjust TPD quota with actual usage
  go func() {
      // ...
      if err := b.quota.AdjustTPD(ctx, uid, totalTokens, reservedTokens); err != nil {
  ```
  - 如果 goroutine 启动前连接断开，`reservedTokens` 不会被释放
  - **修复**: 使用 `defer` 确保配额调整

- [ ] **7. Claude Provider Name 返回硬编码** - `pkg/ai/providers/claude/provider.go:46`
  ```go
  func (p *Provider) Name() string {
      return "claude"
  }
  ```
  - OpenAI Provider 已修复支持 `config.Name`，但 Claude 还没
  - **影响**: 与设计不一致

- [ ] **8. 消息保存时的 model 字段混淆** - `internal/apiserver/biz/chat/chat.go:245`
  ```go
  Model: req.Model,  // 应该是使用的模型
  ```
  - 保存用户消息时用 `req.Model`，但保存助手回复时用 `usedModel`
  - 用户消息不应该关联特定模型，或者应该留空

## Suggestions (可以改进)

- [ ] **9. `GenerateID` 函数重复**
  - 每个 provider 都有自己的 `generateID` 函数，完全相同
  - **建议**: 提取到 `pkg/ai/utils.go`

- [ ] **10. Magic Number**
  - `100` 作为 `NewChatStream` 的 buffer size 出现在多处
  - **建议**: 定义常量 `const defaultStreamBuffer = 100`

- [ ] **11. 错误处理不一致**
  - 有些地方用 `log.C(ctx).Errorw`，有些用 `log.Errorw`
  - **建议**: 统一使用 `log.C(ctx).Errorw` 以便追踪请求上下文

- [ ] **12. 缺少单元测试**
  - `pkg/ai/providers/` 下只有 `openai/provider_test.go`
  - **建议**: 为每个 Provider 添加基本测试

- [ ] **13. `wrapStreamForSaving` 中 contentBuilder 使用 `[]byte`**
  - 但内容是 `string`，频繁类型转换
  - **建议**: 直接使用 `strings.Builder` 或 `[]string`

---

## 做得好的地方 ✅

1. **完整的错误处理** - 所有操作都有适当的错误处理
2. **配额管理使用 Redis 原子操作** - `ReserveTPD` 使用 `INCRBY` 防止竞态
3. **三层架构清晰** - Handler → Biz → Store 职责分明
4. **支持多 Provider** - OpenAI、Claude、Gemini、Qwen
5. **流式响应处理完善** - 正确处理 EOF 和错误情况
6. **会话验证** - 验证 session 属于当前用户，防止越权访问
7. **超时控制** - 后台操作使用 `context.WithTimeout` 防止 goroutine 泄漏

---

## 总结

| 类别 | 数量 | 状态 |
|------|------|------|
| Critical | 3 | ⏳ 待修复 |
| Important | 5 | ⏳ 待修复 |
| Suggestions | 5 | ⏳ 待修复 |

**修复优先级**:
1. #2, #3 - 重复保存消息（数据一致性问题）
2. #1 - Provider 代码重复（维护性问题）
3. #5 - 历史消息可能重复（功能问题）
