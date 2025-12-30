# AI Chat 实现 Review 问题清单

> 2025-12-30 代码审查发现的问题

## Critical (必须修复) ✅ 已全部修复

- [x] **1. AI 限流中间件未应用** - `internal/apiserver/http.go:36-47`
  - `AILimiter` 已定义但未应用到路由
  - **修复**: 在 http.go 中添加 `httpmw.AILimiter(rpm)` 中间件

- [x] **2. goroutine 可能泄漏** - `internal/apiserver/biz/chat/chat.go:63-65`
  - `saveToSession` 用 `context.Background()` 没有超时
  - **修复**: 使用 `context.WithTimeout(context.Background(), 30s)` 并添加错误日志

- [x] **3. TPD Token 配额未实现**
  - 设计文档要求的 `CheckAndUpdateTokenQuota` 完全缺失
  - **修复**: 创建 `quota.go` 实现 CheckTPD/UpdateTPD，在 Chat/ChatStream 中调用

- [x] **4. 流式响应不保存会话** - `internal/apiserver/biz/chat/chat.go:70-88`
  - `ChatStream()` 没有保存消息到 session
  - **修复**: 添加 `wrapStreamForSaving` 包装流，在流结束后保存消息和更新配额

## Important (应该修复) ✅ 已全部修复

- [x] **5. 响应 ID 不唯一** - `pkg/ai/providers/openai/provider.go`
  - `generateID()` 基于秒级时间戳，并发时会重复
  - **修复**: 使用 crypto/rand 生成随机 hex 字符串

- [x] **6. 会话验证缺失** - `internal/apiserver/biz/chat/chat.go`
  - 未校验 session 是否存在或属于当前用户
  - **修复**: 添加 `validateSession` 方法，验证 session 存在且属于当前用户

- [x] **7. 历史消息未使用**
  - 设计的 `BuildMessages` 滑动窗口逻辑未实现
  - **修复**: 添加 `loadAndMergeHistory` 方法，加载历史消息并应用滑动窗口

- [x] **8. Usage 统计空置** - `pkg/ai/providers/openai/provider.go`
  - Eino 不直接暴露 token 计数，Usage 始终为空
  - **修复**: Eino 实际上在 `ResponseMeta.Usage` 中提供了 token 统计
    - `convertResponse` 提取非流式响应的 Usage
    - `convertStreamChunk` 提取流式 chunk 的 Usage
    - 流式 goroutine 追踪 lastUsage 并在 final chunk 中返回

- [x] **9. Stream 缺少 FinishReason** - `pkg/ai/providers/openai/provider.go`
  - 最后一个 chunk 没有设置 finish_reason
  - **修复**: 在流结束时发送带有 finish_reason="stop" 的最终 chunk

- [x] **10. 三层 Model 解析未实现**
  - 设计的 Request > User Preference > System Default 逻辑缺失
  - **修复**: 添加 `resolveModel` 方法实现 请求 > 会话 > 系统默认 的优先级

## Suggestions (可以改进) ✅ 已全部修复

- [x] **11. Provider name 硬编码** - `pkg/ai/providers/openai/provider.go:45`
  - 始终返回 "openai"，不支持 DeepSeek 等兼容服务区分
  - **修复**:
    - Config 添加 Name 字段，Provider.Name() 使用 config.Name
    - http.go 根据 credential name 选择对应的预设配置 (DeepSeekConfig/MoonshotConfig)

- [x] **12. AILimiter 用内存存储** - `internal/pkg/middleware/http/ai_limiter.go:22`
  - 分布式环境不工作，应用 Redis
  - **修复**: 使用 GetLimiterContext 调用 Redis 存储

- [x] **13. 设计中的文件未创建** - 评估后决定不拆分
  - `quota.go` 已创建；`resolveModel`/`loadAndMergeHistory` 保留在 `chat.go` 中
  - 理由：395 行是合理大小，高内聚，拆分会增加复杂度

- [x] **14. PUT sessions/:id 路由未注册**
  - 设计文档要求但未实现
  - **修复**: 添加 UpdateSession handler 和 PUT /:session_id 路由

---

## 总结

| 类别 | 数量 | 状态 |
|------|------|------|
| Critical | 4 | ✅ 全部修复 |
| Important | 6 | ✅ 全部修复 |
| Suggestions | 4 | ✅ 全部修复 |

**相关提交**:
- `1b62e0b` fix(ai): resolve critical issues from code review
- `3d9fce6` fix(ai): resolve important issues from code review
- `f3bc4ed` docs: mark #13 as evaluated - no file split needed
- `40601df` fix(ai): resolve remaining review issues
