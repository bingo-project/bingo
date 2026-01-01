# 通知系统

Bingo 内置了全功能的通知系统，支持站内信、公告、邮件等多种渠道的消息推送。

## 架构概览

通知系统采用 **Redis Pub/Sub** 和 **WebSocket** 实现实时推送，支持 write-diffusion（写扩散）模型用于个人通知，read-diffusion（读扩散）模型用于系统公告。

- **个人通知**：业务触发 -> 写入 `ntf_message` -> Redis 推送 -> WebSocket -> 用户
- **系统公告**：后台发布 -> 写入 `ntf_announcement` -> Redis 广播 -> WebSocket -> 所有在线用户

## 主要功能

### 1. 通知类型 (Category)

- **System**: 系统通知（如维护公告、功能更新）
- **Security**: 安全提醒（如异地登录、密码修改）
- **Transaction**: 交易相关（如充值到账、支付成功）
- **Social**: 社交互动（如被关注、收到评论）

### 2. 多渠道支持

- **In-App**: 站内信（WebSocket 实时推送 + 历史记录）
- **Email**: 邮件通知（基于 Asynq 的异步任务）

用户可以针对不同类型的通知配置接收渠道（例如：接收安全类邮件，但不接收社交类邮件）。

## API 使用指南

### 获取通知列表

支持按分类筛选、按已读状态筛选。列表会自动合并个人通知和系统公告。

```http
GET /v1/notifications?category=security&is_read=false&page=1&page_size=20
```

### 获取未读数

```http
GET /v1/notifications/unread-count
```

### 标记已读

```http
# 标记单条
PUT /v1/notifications/:uuid/read

# 标记全部
PUT /v1/notifications/read-all
```

### 通知偏好设置

用户可以查询和修改自己的通知配置。

```http
GET /v1/notifications/preferences
PUT /v1/notifications/preferences
```

**Payload 示例**:
```json
{
  "transaction": { "in_app": true, "email": true },
  "social": { "in_app": true, "email": false }
}
```

## 服务端调用

业务模块调用 `notification` 包发送通知，系统会自动处理渠道分发和持久化。

```go
import "github.com/bingo-project/bingo/internal/pkg/notification"

// 发送通知
notification.Send(ctx, &notification.Message{
    UserID:    targetUserID,
    Category:  notification.CategorySecurity,
    Type:      "login_alert",
    Title:     "异地登录提醒",
    Content:   "您的账号于 2025-01-01 在新设备登录",
    ActionURL: "/security/logs",
})
```
