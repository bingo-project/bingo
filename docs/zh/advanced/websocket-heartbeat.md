# WebSocket 设计文档

本文档介绍 bingo WebSocket 模块的完整设计，包括认证、心跳、推送和订阅机制。

## 设计目标

1. **连接后认证** - 支持匿名连接，登录后解锁完整功能
2. **多种登录方式** - 支持账号密码登录和 Token 登录（复用 HTTP Token）
3. **单点登录** - 同一平台同一用户只能有一个连接，后登录踢掉前一个
4. **检测死连接** - 及时发现并清理无响应的客户端
5. **保持连接活性** - 防止 NAT 超时、防火墙断开
6. **便捷推送** - 提供友好的服务端推送 API
7. **灵活订阅** - 支持群聊、实时数据等发布/订阅场景

---

## 连接状态机

```
┌─────────────────────────────────────────────────────────────┐
│                    WebSocket 连接状态机                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   ┌───────────┐      login 成功      ┌──────────────┐      │
│   │ Anonymous │ ───────────────────→ │Authenticated │      │
│   │  (匿名)    │                      │   (已认证)    │      │
│   └───────────┘                      └──────────────┘      │
│        │                                    │               │
│        │ 10s 无登录                          │ 60s 无心跳    │
│        ↓                                    ↓               │
│   ┌─────────────────────────────────────────────┐          │
│   │              Disconnected                    │          │
│   └─────────────────────────────────────────────┘          │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│  状态          │ 可用方法                   │ 超时          │
│───────────────┼───────────────────────────┼───────────────│
│  Anonymous    │ login, heartbeat          │ 10s           │
│  Authenticated│ heartbeat, subscribe,     │ 60s           │
│               │ unsubscribe, 业务方法...   │               │
└─────────────────────────────────────────────────────────────┘
```

**关键点**：
- 匿名连接必须在 10 秒内完成登录，否则断开
- 已认证连接使用标准 60 秒心跳超时
- 匿名状态也可以发心跳（保持连接活性，但不能续命超过 10 秒）

---

## 核心概念

### Platform（平台标识）

用于区分不同客户端平台，同一用户可同时在多个平台登录，但不能在同一平台多设备登录（后登录挤掉前一个）。

```go
const (
    PlatformWeb     = "web"      // 网页端
    PlatformIOS     = "ios"      // iOS App
    PlatformAndroid = "android"  // Android App
    PlatformH5      = "h5"       // H5 移动端
    PlatformMiniApp = "miniapp"  // 小程序（微信/支付宝）
    PlatformDesktop = "desktop"  // 桌面端（Electron）
)
```

### 用户标识

- **UserKey**: `platform_userID`，唯一标识一个用户在特定平台的连接
- 例如：用户 `user123` 同时在 iOS 和 Web 登录，会有两个连接：`ios_user123` 和 `web_user123`

### Topic（订阅主题）

用于发布/订阅模式，支持多种实时数据场景。

| 前缀 | 用途 | 示例 |
|-----|------|------|
| `group:` | 群聊 | `group:123` |
| `room:` | 聊天室 | `room:lobby` |
| `doc:` | 协同文档 | `doc:abc123` |
| `board:` | 在线白板 | `board:xyz` |
| `ticker:` | 实时行情 | `ticker:BTC/USDT` |
| `metrics:` | 监控指标 | `metrics:server1` |
| `device:` | IoT 设备 | `device:12345` |

---

## 客户端协议

### 登录

**方式 1：账号密码登录**

```json
{
    "jsonrpc": "2.0",
    "method": "login",
    "params": {
        "type": "password",
        "username": "user@example.com",
        "password": "xxx",
        "platform": "ios"
    },
    "id": 1
}
```

**方式 2：Token 登录（复用 HTTP 登录获取的 token）**

```json
{
    "jsonrpc": "2.0",
    "method": "login",
    "params": {
        "type": "token",
        "token": "eyJhbGciOiJIUzI1NiIs..."
    },
    "id": 2
}
```

**成功响应**：

```json
{
    "jsonrpc": "2.0",
    "result": {
        "user_id": "123",
        "platform": "ios",
        "token": "eyJhbGciOiJIUzI1NiIs...",
        "expires_at": 1702396800
    },
    "id": 1
}
```

**失败响应**：

```json
{
    "jsonrpc": "2.0",
    "error": {
        "code": -32001,
        "message": "Invalid credentials"
    },
    "id": 1
}
```

**关键点**：
- 密码登录返回新 token（和 HTTP 登录一致）
- Token 登录也返回 token（原样返回或刷新，取决于实现）
- `expires_at` 让客户端知道何时需要重连
- Platform 来源：密码登录从 params 获取，token 登录从 token payload 获取

### 心跳

**请求**：
```json
{
    "jsonrpc": "2.0",
    "method": "heartbeat",
    "id": 1
}
```

**响应**：
```json
{
    "jsonrpc": "2.0",
    "result": {
        "status": "ok",
        "server_time": 1701792000
    },
    "id": 1
}
```

### 订阅 Topic

**订阅请求**：
```json
{
    "jsonrpc": "2.0",
    "method": "subscribe",
    "params": {
        "topics": ["group:123", "room:lobby"]
    },
    "id": 2
}
```

**订阅响应**：
```json
{
    "jsonrpc": "2.0",
    "result": {
        "subscribed": ["group:123", "room:lobby"]
    },
    "id": 2
}
```

### 取消订阅

**请求**：
```json
{
    "jsonrpc": "2.0",
    "method": "unsubscribe",
    "params": {
        "topics": ["group:123"]
    },
    "id": 3
}
```

**响应**：
```json
{
    "jsonrpc": "2.0",
    "result": {
        "unsubscribed": ["group:123"]
    },
    "id": 3
}
```

### 客户端要求

1. 连接后 **10 秒内** 必须完成登录
2. 登录后 **每 30 秒** 发送一次 `heartbeat` 请求
3. **60 秒** 内无任何消息将被服务端断开
4. 断开后需自行实现 **重连逻辑**
5. 重连后需 **重新登录** 并 **重新订阅** Topic

---

## 双层心跳架构

```
┌─────────────────────────────────────────────────────────────┐
│                    协议层 (WebSocket)                        │
│                                                              │
│  服务端 ──── ping (每54s) ────→ 客户端                       │
│  服务端 ←─── pong ─────────── 客户端                        │
│                                                              │
│  目的：检测 TCP 连接活性                                      │
│  超时：60s 未收到 pong → 断开                                │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    应用层 (JSON-RPC)                         │
│                                                              │
│  服务端 ←─── heartbeat (每30s) ─── 客户端                   │
│  服务端 ──── response ──────────→ 客户端                    │
│                                                              │
│  目的：                                                      │
│    1. 确认客户端还在消费数据                                  │
│    2. 保持 NAT 映射                                          │
│    3. 客户端确认服务端活着                                    │
│  超时：60s 未收到任何消息 → 断开                             │
└─────────────────────────────────────────────────────────────┘
```

### 为什么需要双层心跳？

| 层级 | 检测目标 | 场景 |
|-----|---------|------|
| **协议层** | TCP 连接是否存活 | 网络断开、客户端崩溃 |
| **应用层** | 客户端是否在消费数据 | 客户端进程假死、App 被杀但 TCP 未断 |

**典型问题场景**：服务端持续推送数据，客户端 App 被杀掉，但 TCP 连接因为操作系统缓冲区未满而保持"活着"。协议层 ping/pong 正常，但数据没人消费。应用层心跳能发现这种情况。

---

## 单点登录踢人机制

**场景**：用户在 iOS 设备 A 已登录，又在 iOS 设备 B 登录

**流程**：

```
设备 B 发送 login 请求
        ↓
服务端验证成功
        ↓
检查 users["ios_123"] 是否存在
        ↓ 存在
向设备 A 发送踢人通知
        ↓
{
    "jsonrpc": "2.0",
    "method": "session.kicked",
    "params": {
        "reason": "您的账号已在其他设备登录"
    }
}
        ↓
100ms 后断开设备 A 连接
        ↓
更新 users["ios_123"] = 设备 B
        ↓
向设备 B 返回登录成功
```

**实现**：

```go
func (h *Hub) handleLogin(event *LoginEvent) {
    userKey := event.Platform + "_" + event.UserID

    h.userLock.Lock()
    oldClient := h.users[userKey]
    h.users[userKey] = event.Client
    h.userLock.Unlock()

    if oldClient != nil {
        // 先通知，再踢
        oldClient.Send <- buildKickNotification("您的账号已在其他设备登录")
        time.AfterFunc(100*time.Millisecond, func() {
            h.Unregister <- oldClient
        })
    }
}
```

---

## Token 过期处理

**场景**：用户已认证的 WebSocket 连接，token 过期了

**方案**：服务端主动通知并断开

```
服务端检测 token 过期
        ↓
发送通知
{
    "jsonrpc": "2.0",
    "method": "session.expired",
    "params": {
        "reason": "Token 已过期，请重新登录"
    }
}
        ↓
100ms 后断开连接
```

**客户端处理**：

1. 收到 `session.expired` 通知
2. 连接断开
3. 自动重连
4. 使用 refresh_token 调用 HTTP 接口获取新 token（或重新登录）
5. 用新 token 调用 WebSocket login

---

## 服务端推送

### 消息格式

服务端主动推送使用 JSON-RPC 2.0 Notification 格式（无 `id` 字段）：

```json
{
    "jsonrpc": "2.0",
    "method": "message.new",
    "params": {
        "group_id": "123",
        "content": "Hello!",
        "sender_id": "456",
        "timestamp": 1701792000
    }
}
```

### 推送 API

```go
// ==================== 用户推送 ====================

// PushToUser 向指定平台的用户推送
hub.PushToUser("ios", "user123", "order.created", data)

// PushToUserAllPlatforms 向用户的所有平台推送
hub.PushToUserAllPlatforms("user123", "security.alert", data)

// PushToUsers 向指定平台的多个用户推送
hub.PushToUsers("ios", []string{"user1", "user2"}, "promo.new", data)

// ==================== Topic 推送 ====================

// PushToTopic 向订阅了指定 Topic 的所有客户端推送
hub.PushToTopic("group:123", "message.new", data)

// PushToTopics 向多个 Topic 推送（去重）
hub.PushToTopics([]string{"group:1", "group:2"}, "message.new", data)

// ==================== 广播 ====================

// BroadcastToPlatform 向指定平台的所有连接广播
hub.BroadcastToPlatform("ios", "app.update", data)

// Broadcast 向所有连接广播
hub.Broadcast("system.maintenance", data)
```

### 常见推送场景

| 场景 | method | 推送方式 |
|-----|--------|---------|
| 订单通知 | `order.created` | `PushToUser` |
| 账户安全 | `security.alert` | `PushToUserAllPlatforms` |
| 群聊消息 | `message.new` | `PushToTopic("group:xxx")` |
| App 更新 | `app.update` | `BroadcastToPlatform` |
| 系统维护 | `system.maintenance` | `Broadcast` |

---

## Topic 订阅机制

**设计**：通过 channel 串行化，所有订阅操作都由 Hub.Run 处理，避免死锁

### 事件定义

```go
type SubscribeEvent struct {
    Client *Client
    Topics []string
    Result chan []string  // 返回成功订阅的 topics
}

type UnsubscribeEvent struct {
    Client *Client
    Topics []string
}
```

### Hub.Run 处理

```go
case event := <-h.Subscribe:
    subscribed := h.doSubscribe(event.Client, event.Topics)
    if event.Result != nil {
        event.Result <- subscribed
    }

case event := <-h.Unsubscribe:
    h.doUnsubscribe(event.Client, event.Topics)
```

### doSubscribe 实现

```go
// 无锁，因为只在 Run goroutine 中执行
func (h *Hub) doSubscribe(client *Client, topics []string) []string {
    var subscribed []string
    for _, topic := range topics {
        if h.topics[topic] == nil {
            h.topics[topic] = make(map[*Client]bool)
        }
        h.topics[topic][client] = true
        client.topics[topic] = true
        subscribed = append(subscribed, topic)
    }
    return subscribed
}
```

---

## 群组成员变动的实时订阅

**场景**：用户在线时被踢出群组，或加入新群组

### 事件定义

```go
type GroupMembershipEvent struct {
    Type    string // "join" | "leave"
    UserID  string
    GroupID string
}
```

### Biz 层调用

```go
// 添加成员
func (b *groupBiz) AddMember(ctx context.Context, groupID, userID string) error {
    if err := b.store.GroupMember().Add(ctx, groupID, userID); err != nil {
        return err
    }

    b.hub.GroupMembership <- &GroupMembershipEvent{
        Type:    "join",
        UserID:  userID,
        GroupID: groupID,
    }
    return nil
}

// 移除成员
func (b *groupBiz) RemoveMember(ctx context.Context, groupID, userID string) error {
    if err := b.store.GroupMember().Remove(ctx, groupID, userID); err != nil {
        return err
    }

    b.hub.GroupMembership <- &GroupMembershipEvent{
        Type:    "leave",
        UserID:  userID,
        GroupID: groupID,
    }
    return nil
}
```

---

## 数据结构

### Hub 结构

```go
type Hub struct {
    config HubConfig

    // 匿名连接（未登录）
    anonymous     map[*Client]bool
    anonymousLock sync.RWMutex

    // 已认证连接
    clients     map[*Client]bool
    clientsLock sync.RWMutex

    // 按 platform_userID 索引
    users    map[string]*Client
    userLock sync.RWMutex

    // Topic 订阅关系
    topics map[string]map[*Client]bool  // topic -> clients

    // 事件通道
    Register        chan *Client
    Unregister      chan *Client
    Login           chan *LoginEvent
    Subscribe       chan *SubscribeEvent
    Unsubscribe     chan *UnsubscribeEvent
    GroupMembership chan *GroupMembershipEvent
    Broadcast       chan []byte
}
```

### Client 结构

```go
type Client struct {
    hub     *Hub
    conn    *websocket.Conn
    adapter *jsonrpc.Adapter
    ctx     context.Context

    Send chan []byte

    // 连接信息
    Addr           string
    Platform       string
    UserID         string
    FirstTime      int64
    HeartbeatTime  int64
    LoginTime      int64
    TokenExpiresAt int64

    // 订阅的 Topics（由 Hub 管理，Client 只读）
    topics map[string]bool
}
```

---

## 参数配置

```go
type HubConfig struct {
    // 匿名连接
    AnonymousTimeout time.Duration  // 默认 10s
    AnonymousCleanup time.Duration  // 默认 2s

    // 已认证连接
    HeartbeatTimeout time.Duration  // 默认 60s
    HeartbeatCleanup time.Duration  // 默认 30s

    // WebSocket 协议层
    PingPeriod time.Duration        // 默认 54s
    PongWait   time.Duration        // 默认 60s
}
```

| 参数 | 值 | 说明 |
|-----|-----|-----|
| `AnonymousTimeout` | 10s | 匿名连接必须在此时间内登录 |
| `AnonymousCleanup` | 2s | 匿名连接扫描间隔 |
| `HeartbeatTimeout` | 60s | 已认证连接心跳超时 |
| `HeartbeatCleanup` | 30s | 已认证连接扫描间隔 |
| `HeartbeatInterval` | 30s | 建议客户端发心跳间隔（文档告知） |
| `PingPeriod` | 54s | 服务端 WebSocket ping 间隔 |
| `PongWait` | 60s | 等待 pong 超时 |

---

## 适用场景

bingo WebSocket 模块设计为通用高性能方案，支持以下典型场景开箱即用：

### 实时协作

| 场景 | Topic 示例 | 消息特点 |
|-----|-----------|---------|
| 协同文档 | `doc:{docID}` | 多人编辑、光标同步、操作变换 |
| 在线白板 | `board:{boardID}` | 图形操作、实时同步 |
| 代码协作 | `code:{sessionID}` | 代码变更、终端输出 |

### 即时通讯

| 场景 | Topic 示例 | 消息特点 |
|-----|-----------|---------|
| 私聊 | 点对点推送 | 已读回执、撤回通知 |
| 群聊 | `group:{groupID}` | 广播、@提醒 |
| 聊天室 | `room:{roomID}` | 大规模广播、弹幕 |

### 数据推送

| 场景 | Topic 示例 | 消息特点 |
|-----|-----------|---------|
| 实时行情 | `ticker:{symbol}` | 高频小消息、10-100次/秒 |
| 订单状态 | 用户私有推送 | 状态变更通知 |
| 体育比分 | `match:{matchID}` | 事件驱动 |

### 监控运维

| 场景 | Topic 示例 | 消息特点 |
|-----|-----------|---------|
| 服务器监控 | `metrics:{serverID}` | CPU/内存/网络指标 |
| 日志流 | `logs:{serviceID}` | 持续推送、类似 tail -f |
| CI/CD 构建 | `build:{buildID}` | 构建日志实时输出 |

### IoT 设备

| 场景 | Topic 示例 | 消息特点 |
|-----|-----------|---------|
| 智能家居 | `device:{deviceID}` | 状态上报、指令下发 |
| 位置追踪 | `location:{entityID}` | 骑手/车辆实时位置 |

---

## 数据获取策略

### 历史数据 vs 实时数据

| 数据类型 | 接口 | 压缩方式 | 说明 |
|---------|------|---------|------|
| 历史记录 | REST API | HTTP gzip | 聊天记录、操作日志等 |
| 批量数据 | REST API | HTTP gzip | 列表、统计报表等 |
| 实时更新 | WebSocket | 不压缩 | 单条消息 < 200 字节 |
| 状态变更 | WebSocket | 不压缩 | 通知、提醒等 |

**设计原则**：
- REST API 用于拉取历史/批量数据，利用 HTTP 标准 gzip
- WebSocket 只用于实时推送，消息保持精简

### REST API 压缩

使用 Gin gzip middleware，客户端无需额外处理：

```go
import "github.com/gin-contrib/gzip"

r.Use(gzip.Gzip(gzip.DefaultCompression))
```

### WebSocket 暂不压缩

原因：
- 实时消息本身小（< 200 字节），压缩收益低
- 高频场景 CPU 是瓶颈，压缩增加开销
- 如未来有大消息场景，可按消息大小决定是否压缩

---

## 性能边界与扩展

### 适用规模

当前设计适用于中小型场景：

| 指标 | 建议上限 | 说明 |
|-----|---------|------|
| 单机连接数 | 5 万 | 受内存和 CPU 限制 |
| 单 Topic 订阅者 | 1000 | 广播遍历开销 |
| 消息频率 | 100 次/秒/Topic | 写入吞吐限制 |

超过此规模需要考虑优化或分布式方案。

### 扩展点预留

**写合并（Batch Write）**：

高频场景（弹幕、行情）可启用写合并，减少系统调用：

```go
// MessageWriter 接口预留
type MessageWriter interface {
    Write(msg []byte) error
    Flush() error
}

// 默认：DirectWriter（立即写入）
// 高频：BatchWriter（50ms 合并写入）
```

启用写合并的代价：
- 延迟增加 ~50ms
- 客户端需处理 NDJSON 格式（多条消息换行分隔）
- 连接断开时 buffer 中消息可能丢失

**并行广播（Fan-out Workers）**：

大 Topic（> 1000 订阅者）可启用并行广播：

```go
// 将订阅者分片，多 goroutine 并行推送
func (h *Hub) PushToTopicParallel(topic string, msg []byte, workers int)
```

**分布式扩展**：

超过单机容量时，通过 Redis Pub/Sub 实现多实例：

```
用户 → 实例 A ←→ Redis Pub/Sub ←→ 实例 B ← 用户
```

---

## 文件变更清单

| 文件 | 变更类型 | 说明 |
|-----|---------|------|
| `pkg/ws/hub.go` | 修改 | 新增 anonymous map、事件 channel、清理任务拆分 |
| `pkg/ws/client.go` | 修改 | 新增 TokenExpiresAt、状态判断方法 |
| `pkg/ws/handler.go` | 修改 | 新增 login 方法处理、状态检查中间件 |
| `pkg/ws/config.go` | 修改 | 新增 AnonymousTimeout/Cleanup 配置 |

---

## 相关文档

- [WebSocket 统一方案](websocket-unification.md) - JSON-RPC 2.0 消息格式
- [可插拔协议层](protocol-layer.md) - 多协议架构设计
