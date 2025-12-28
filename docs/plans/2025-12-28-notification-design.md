# 通知系统设计

## 概述

为 Bingo 项目增加通知功能，支持：
- 用户通知偏好设置（按类型 + 渠道控制）
- 通知中心（历史通知列表、已读、删除、筛选）
- 公告发布（立即推送 + 定时推送）
- 实时推送（WebSocket）

## 核心决策

| 项目 | 决策 |
|------|------|
| 服务间通信 | Redis Pub/Sub（推送）+ gRPC（查询在线用户） |
| 存储策略 | 个人通知写扩散 + 公告单独存储 + 已读关联表 |
| 通知偏好 | 按类型 + 渠道，JSON 存储 |
| 通知类型 | 系统、安全、交易、社交 |
| 推送渠道 | 站内 + 邮件（短信、App Push 预留） |
| 公告发布 | 立即 + 定时（Asynq） |
| 任务处理 | 统一在 scheduler |
| 公告推送 | 全员广播 |

## 数据模型

### ntf_message（个人通知表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 主键 |
| uuid | string | 唯一标识，对外暴露 |
| user_id | string | 接收用户，关联 user.uid |
| category | string | 大类：system/security/transaction/social |
| type | string | 具体类型：login_alert, deposit_success 等 |
| title | string | 标题 |
| content | string | 内容 |
| action_url | string | 详情跳转链接，可空 |
| is_read | bool | 是否已读 |
| read_at | timestamp | 已读时间 |
| created_at | timestamp | 创建时间 |

### ntf_announcement（公告表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 主键 |
| uuid | string | 唯一标识 |
| title | string | 标题 |
| content | string | 内容 |
| action_url | string | 跳转链接，可空 |
| status | string | 状态：draft/scheduled/published |
| scheduled_at | timestamp | 定时发布时间 |
| published_at | timestamp | 实际发布时间 |
| expires_at | timestamp | 过期时间 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### ntf_announcement_read（公告已读表）

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | string | 用户 uid |
| announcement_id | uint64 | 公告 id |
| read_at | timestamp | 已读时间 |

### ntf_preference（用户通知偏好表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 主键 |
| user_id | string | 用户 uid，唯一 |
| preferences | json | 各类型各渠道的开关 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

**preferences JSON 结构**：

```json
{
  "system":      { "in_app": true, "email": false },
  "security":    { "in_app": true, "email": true  },
  "transaction": { "in_app": true, "email": true  },
  "social":      { "in_app": true, "email": false }
}
```

**默认配置**（代码中定义）：

| 类型 | 站内 | 邮件 |
|------|------|------|
| system | ✓ | ✗ |
| security | ✓ | ✓ |
| transaction | ✓ | ✓ |
| social | ✓ | ✗ |

## 服务间通信

### 架构图

```
┌─────────────┐     Redis Pub/Sub      ┌─────────────┐     WebSocket      ┌────────┐
│  admserver  │ ──────────────────────▶│  apiserver  │ ──────────────────▶│  用户  │
└─────────────┘   ntf:broadcast         └─────────────┘                    └────────┘
       │                                       │
       │          gRPC                         │
       └──────────────────────────────────────▶│
              查询在线用户
```

### Redis Pub/Sub 频道

| 频道 | 用途 | 消息方向 |
|------|------|----------|
| `ntf:broadcast` | 公告推送 | admserver → apiserver |
| `ntf:user:{user_id}` | 个人通知推送 | 业务服务 → apiserver |

### WebSocket 推送格式（JSON-RPC）

```json
{
  "jsonrpc": "2.0",
  "method": "ntf.announcement",
  "data": {
    "uuid": "xxx",
    "title": "系统升级通知",
    "content": "...",
    "action_url": "/announcements/xxx"
  }
}
```

**推送 method 命名**：

| method | 用途 |
|--------|------|
| `ntf.announcement` | 公告推送 |
| `ntf.message` | 个人通知推送 |
| `ntf.unread_count` | 未读数变更推送 |

### gRPC 接口（apiserver 提供）

```protobuf
service WSService {
  rpc GetOnlineUsers(Empty) returns (OnlineUsersResponse);
  rpc GetOnlineCount(Empty) returns (OnlineCountResponse);
  rpc IsUserOnline(UserRequest) returns (BoolResponse);
}
```

## API 设计

### apiserver - 通知中心

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/v1/notifications` | 通知列表（分页、按类型筛选） |
| GET | `/v1/notifications/unread-count` | 获取未读数 |
| PUT | `/v1/notifications/:uuid/read` | 标记单条已读 |
| PUT | `/v1/notifications/read-all` | 标记全部已读 |
| DELETE | `/v1/notifications/:uuid` | 删除单条通知 |
| GET | `/v1/notifications/preferences` | 获取通知偏好 |
| PUT | `/v1/notifications/preferences` | 更新通知偏好 |

**列表接口参数**：

```
GET /v1/notifications?category=security&is_read=false&page=1&page_size=20
```

**列表响应**（合并个人通知 + 公告）：

```json
{
  "data": [
    {
      "uuid": "xxx",
      "source": "message",
      "category": "security",
      "type": "login_alert",
      "title": "异地登录提醒",
      "content": "...",
      "action_url": "/security/logs",
      "is_read": false,
      "created_at": "2025-12-28T10:00:00Z"
    },
    {
      "uuid": "yyy",
      "source": "announcement",
      "category": "system",
      "title": "系统升级通知",
      "content": "...",
      "is_read": true,
      "created_at": "2025-12-27T10:00:00Z"
    }
  ],
  "total": 50
}
```

**说明**：
- `source` 字段区分来源（message/announcement）
- 公告不支持删除，只支持标记已读
- 列表按 `created_at` 倒序排列

### admserver - 公告管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/v1/announcements` | 公告列表（分页、按状态筛选） |
| GET | `/v1/announcements/:uuid` | 公告详情 |
| POST | `/v1/announcements` | 创建公告（草稿） |
| PUT | `/v1/announcements/:uuid` | 编辑公告 |
| DELETE | `/v1/announcements/:uuid` | 删除公告（仅草稿可删） |
| POST | `/v1/announcements/:uuid/publish` | 立即发布 |
| POST | `/v1/announcements/:uuid/schedule` | 定时发布 |
| POST | `/v1/announcements/:uuid/cancel` | 取消定时 |

**公告状态流转**：

```
draft ──▶ scheduled ──▶ published
  │            │
  │            ▼
  │         cancelled (取消定时，回到 draft)
  ▼
published (立即发布)
```

## 推送流程

### 定时发布流程

```
┌─────────────┐    创建延时任务     ┌─────────┐    到点执行     ┌─────────────┐
│  admserver  │ ─────────────────▶ │  Asynq  │ ─────────────▶ │  scheduler  │
│  设置定时   │                     │  队列   │                │  Worker     │
└─────────────┘                     └─────────┘                └──────┬──────┘
                                                                      │
                                                                      ▼
                                                               1. 更新状态为 published
                                                               2. 发布到 Redis Pub/Sub
                                                                      │
                                                                      ▼
                                                               ┌─────────────┐
                                                               │  apiserver  │
                                                               │  subscriber │
                                                               └──────┬──────┘
                                                                      │
                                                                      ▼
                                                               hub.Broadcast
                                                               推送给所有用户
```

### 个人通知触发流程

```
┌──────────────┐    业务事件     ┌─────────────┐    Redis Pub/Sub    ┌─────────────┐
│  业务代码    │ ─────────────▶ │  通知服务   │ ─────────────────▶  │  apiserver  │
│  (biz层)     │                │  (封装)     │  ntf:user:{user_id} └──────┬──────┘
└──────────────┘                └─────────────┘                            │
                                       │                                   ▼
                                       ▼                            WebSocket 推送
                                写入 ntf_message 表                   给在线用户
```

**通知服务调用示例**：

```go
notification.Send(ctx, &notification.Message{
    UserID:    "user_uid",
    Category:  notification.CategorySecurity,
    Type:      "login_alert",
    Title:     "异地登录提醒",
    Content:   "您的账号在新设备登录",
    ActionURL: "/security/logs",
})
```

**Send 方法内部逻辑**：

1. 检查用户通知偏好，判断是否发送
2. 写入 `ntf_message` 表
3. 站内渠道：发布到 Redis `ntf:user:{user_id}`
4. 邮件渠道：入队 Asynq 异步发送

## 目录结构

```
internal/
├── apiserver/
│   ├── biz/
│   │   └── notification/
│   │       ├── notification.go       # 通知列表、已读、删除
│   │       └── preference.go         # 偏好设置
│   ├── handler/
│   │   └── http/
│   │       └── notification/
│   │           ├── notification.go
│   │           └── preference.go
│   └── subscriber/
│       └── notification.go           # 订阅 Redis，推送给用户
│
├── admserver/
│   ├── biz/
│   │   └── notification/
│   │       └── announcement.go       # 公告 CRUD、入队发布任务
│   └── handler/
│       └── http/
│           └── notification/
│               └── announcement.go
│
├── scheduler/
│   └── job/
│       ├── registry.go
│       ├── email_verification.go     # 现有
│       └── announcement_publish.go   # 定时发布公告
│
└── pkg/
    ├── notification/                 # 通知发送封装
    │   ├── notification.go           # Send 方法
    │   ├── category.go               # 类型常量
    │   └── channel.go                # 渠道常量
    ├── task/
    │   └── announcement.go           # 任务类型、Payload
    └── model/
        └── notification/
            ├── ntf_message.go
            ├── ntf_announcement.go
            ├── ntf_announcement_read.go
            └── ntf_preference.go
```

## 待补充到 CONVENTIONS.md

### 表名规范

- 统一使用单数形式
- 使用模块前缀，如 `ntf_`、`sys_`
- 示例：`ntf_message`、`ntf_announcement`
