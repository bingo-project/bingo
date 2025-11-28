# Bot 机器人服务

Bingo Bot 是一个多平台机器人服务，支持 Telegram 和 Discord，提供服务管理、通知推送和频道订阅等功能。

## 核心特性

- **多平台支持** - 同时支持 Telegram 和 Discord
- **服务管理** - 健康检查、版本查询、维护模式切换
- **通知订阅** - 频道订阅管理，支持推送系统通知
- **统一接口** - 两个平台使用相同的业务逻辑和数据存储

## 支持的平台

### Telegram
基于 [telebot.v3](https://github.com/tucnak/telebot) 实现，支持：
- Bot 命令交互
- 频道消息推送
- 群组管理

### Discord
基于 [discordgo](https://github.com/bwmarrin/discordgo) 实现，支持：
- Slash Commands
- 频道消息推送
- 服务器管理

## 快速开始

### 1. 创建 Bot

#### Telegram Bot

1. 在 Telegram 中找到 [@BotFather](https://t.me/botfather)
2. 发送 `/newbot` 创建新机器人
3. 按提示设置机器人名称和用户名
4. 获取 Bot Token（请妥善保管，不要泄露）

#### Discord Bot

1. 访问 [Discord Developer Portal](https://discord.com/developers/applications)
2. 点击 "New Application" 创建应用
3. 进入 "Bot" 标签，点击 "Add Bot"
4. 在 "TOKEN" 部分获取 Bot Token
5. 启用必要的 Intents（如 Message Content Intent）

### 2. 配置文件

创建 `bingo-bot.yaml` 配置文件：

```yaml
# Bot server
server:
  name: bingo-bot
  mode: release
  addr: :18080
  timezone: Asia/Shanghai
  key: your-secret-key

# Bot 配置
bot:
  telegram: "YOUR_TELEGRAM_BOT_TOKEN"  # Telegram Bot Token
  discord: "YOUR_DISCORD_BOT_TOKEN"    # Discord Bot Token

# MySQL 配置
mysql:
  host: mysql:3306
  username: root
  password: root
  database: bingo
  maxIdleConnections: 100
  maxOpenConnections: 100
  maxConnectionLifeTime: 10s
  logLevel: 4

# Redis 配置
redis:
  host: redis:6379
  password: ""
  database: 1

# JWT 配置
jwt:
  secretKey: your-jwt-secret-key
  ttl: 1440  # token 过期时间(分钟)

# 日志配置
log:
  level: info
  days: 7
  format: console
  console: true
  maxSize: 100
  compress: true
  path: storage/log/bot.log

feature:
  profiling: true  # 性能分析

# 邮件服务
mail:
  host: "smtp.example.com"
  port: 465
  username: "bot@example.com"
  password: "your-password"
  fromAddr: "noreply@example.com"
  fromName: "Bingo Bot"

# 验证码配置
code:
  length: 6
  ttl: 5        # 有效期（分钟）
  waiting: 1    # 重发等待时间（分钟）
```

**⚠️ 安全提醒：**
- 请将 `YOUR_TELEGRAM_BOT_TOKEN` 和 `YOUR_DISCORD_BOT_TOKEN` 替换为你从 BotFather 和 Discord Developer Portal 获取的真实 Token
- **切勿将包含真实 Token 的配置文件提交到 Git 仓库或公开分享**
- 建议使用环境变量或密钥管理服务来存储 Token
- 如果 Token 泄露，请立即在相应平台重新生成新的 Token

### 3. 启动服务

```bash
# 使用默认配置
./bingo-bot

# 指定配置文件
./bingo-bot -c /path/to/bingo-bot.yaml
```

### 4. 邀请 Bot 到频道

#### Telegram

1. 将 Bot 添加到群组或频道
2. 给予 Bot 必要的权限（发送消息、管理消息等）

#### Discord

1. 在 Developer Portal 的 "OAuth2" > "URL Generator" 中：
   - Scopes 选择：`bot`、`applications.commands`
   - Bot Permissions 选择必要权限
2. 复制生成的 URL，在浏览器中打开
3. 选择服务器并授权

## 可用命令

### 服务管理命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `/ping` 或 `/pong` | 健康检查 | `/ping` → `pong` |
| `/healthz` | 服务状态检查 | `/healthz` → `ok` |
| `/version` | 查看服务版本 | `/version` → `v1.0.0` |
| `/maintenance` | 切换维护模式 | `/maintenance` → `Operation success` |

### 订阅管理命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `/subscribe` | 订阅系统通知 | `/subscribe` → `Successfully subscribe` |
| `/unsubscribe` | 取消订阅 | `/unsubscribe` → `Successfully unsubscribe` |

## 使用示例

### Telegram Bot 交互

```
用户: /ping
Bot:  pong

用户: /version
Bot:  v1.0.0

用户: /subscribe
Bot:  Successfully subscribe, enjoy it!

用户: /healthz
Bot:  ok
```

### Discord Bot 交互（Slash Commands）

```
/ping
Bot: pong

/subscribe
Bot: Successfully subscribe, enjoy it!
```

## 通知推送

Bot 服务支持向已订阅的频道推送系统通知。

### 数据库结构

订阅信息存储在 `bot_channels` 表：

```sql
CREATE TABLE `bot_channels` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `source` varchar(20) NOT NULL COMMENT '平台来源: telegram, discord',
  `channel_id` varchar(100) NOT NULL COMMENT '频道ID',
  `author` text COMMENT '订阅者信息（JSON）',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态: 1-启用, 0-禁用',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_source_channel` (`source`, `channel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 发送通知

在其他服务中通过 API 或数据库触发通知：

```go
import (
    "bingo/internal/pkg/store"
    "bingo/internal/pkg/model/bot"
)

// 获取所有已订阅的频道
channels, err := store.S.Channels().List(ctx, &bot.ListChannelsOptions{
    Status: bot.StatusEnabled,
})

// 向每个频道发送消息
for _, channel := range channels {
    sendMessage(channel.Source, channel.ChannelID, "系统通知: 服务已更新")
}
```

## 架构设计

### 分层架构

```
┌─────────────────────────────────────┐
│  Platform Layer (Telegram/Discord) │  ← 平台适配层
├─────────────────────────────────────┤
│     Controller Layer                │  ← 控制器层
├─────────────────────────────────────┤
│     Business Logic (Biz)            │  ← 业务逻辑层
├─────────────────────────────────────┤
│     Data Access (Store)             │  ← 数据访问层
└─────────────────────────────────────┘
```

### 代码结构

```
internal/bot/
├── biz/               # 业务逻辑层
│   ├── bot/          # Bot 相关业务
│   └── syscfg/       # 系统配置
├── telegram/         # Telegram 平台
│   ├── controller/   # 控制器
│   ├── middleware/   # 中间件
│   ├── router.go     # 路由配置
│   └── run.go        # 启动入口
├── discord/          # Discord 平台
│   ├── controller/   # 控制器
│   ├── middleware/   # 中间件
│   ├── client/       # 客户端封装
│   ├── router.go     # 路由配置
│   └── run.go        # 启动入口
├── app.go            # 应用入口
└── run.go            # 主运行函数
```

## 高级功能

### 自定义命令

#### 1. 添加新命令处理器

```go
// internal/bot/telegram/controller/v1/custom/custom.go
package custom

import (
    "gopkg.in/telebot.v3"
)

type CustomController struct {
    // ...
}

func (ctrl *CustomController) Hello(c telebot.Context) error {
    return c.Send("Hello, " + c.Sender().FirstName + "!")
}
```

#### 2. 注册路由

```go
// internal/bot/telegram/router.go
func RegisterRouter(b *telebot.Bot) {
    // ... 其他路由

    // 自定义命令
    customCtrl := custom.New(store.S)
    b.Handle("/hello", customCtrl.Hello)
}
```

### 中间件

Bot 服务支持中间件来处理通用逻辑：

```go
// 日志中间件示例
func Logger(next telebot.HandlerFunc) telebot.HandlerFunc {
    return func(c telebot.Context) error {
        start := time.Now()

        // 执行处理器
        err := next(c)

        // 记录日志
        log.Infof("Command: %s, User: %s, Duration: %v",
            c.Text(), c.Sender().Username, time.Since(start))

        return err
    }
}

// 注册中间件
b.Use(Logger)
```

### 权限控制

```go
// 管理员权限中间件
func AdminOnly(next telebot.HandlerFunc) telebot.HandlerFunc {
    return func(c telebot.Context) error {
        // 检查用户是否为管理员
        if !isAdmin(c.Sender().ID) {
            return c.Send("此命令仅管理员可用")
        }

        return next(c)
    }
}

// 使用权限中间件
b.Handle("/maintenance", ctrl.ToggleMaintenance, AdminOnly)
```

## 运维监控

### 日志查看

```bash
# 实时查看日志
tail -f storage/log/bot.log

# 查看错误日志
grep "ERROR" storage/log/bot.log

# 查看特定命令日志
grep "/subscribe" storage/log/bot.log
```

### 性能分析

如果启用了 `profiling`，可以访问性能分析接口：

```bash
# 查看堆内存
curl http://localhost:18080/debug/pprof/heap

# 查看 Goroutine
curl http://localhost:18080/debug/pprof/goroutine

# 生成 CPU Profile（30秒）
curl http://localhost:18080/debug/pprof/profile?seconds=30 > cpu.prof
```

### 常见问题

#### 1. Bot 无响应

**检查项：**
- Bot Token 是否正确
- 网络连接是否正常
- Bot 是否被添加到频道
- Bot 是否有足够的权限

```bash
# 检查服务状态
ps aux | grep bingo-bot

# 查看日志
tail -100 storage/log/bot.log
```

#### 2. 订阅失败

**可能原因：**
- 数据库连接失败
- 频道 ID 重复
- 权限不足

**解决方案：**
```bash
# 检查数据库
mysql -h mysql -u root -p bingo
SELECT * FROM bot_channels;

# 检查日志
grep "Subscribe" storage/log/bot.log
```

#### 3. 消息发送失败

**检查：**
- Bot 是否在频道中
- Bot 权限是否足够
- 频道 ID 是否正确
- API 限流（Telegram: 30 msg/sec，Discord: 5 req/sec）

## 最佳实践

### 1. 错误处理

```go
func (ctrl *ServerController) SomeCommand(c telebot.Context) error {
    // 详细的错误信息
    result, err := ctrl.b.DoSomething(ctx)
    if err != nil {
        log.Errorf("操作失败: %v", err)
        return c.Send("操作失败，请稍后重试")
    }

    return c.Send(fmt.Sprintf("操作成功: %s", result))
}
```

### 2. 消息格式化

```go
// Telegram 支持 Markdown 和 HTML
func (ctrl *ServerController) Status(c telebot.Context) error {
    message := `
*服务状态*
━━━━━━━━━━━━
🟢 API Server: 运行中
🟢 Scheduler: 运行中
🟡 Bot: 维护中
━━━━━━━━━━━━
更新时间: %s
    `

    return c.Send(fmt.Sprintf(message, time.Now().Format("15:04:05")),
        &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}
```

### 3. 频率限制

```go
// 使用限流器防止滥用
var limiter = rate.NewLimiter(rate.Every(time.Second), 10)

func (ctrl *ServerController) RateLimitedCommand(c telebot.Context) error {
    if !limiter.Allow() {
        return c.Send("请求过于频繁，请稍后再试")
    }

    // 处理命令
    return nil
}
```

### 4. 定时任务集成

结合 Scheduler 服务发送定时通知：

```go
// 在 Scheduler 中定义任务
func SendDailyReport(ctx context.Context, t *asynq.Task) error {
    // 生成报告
    report := generateReport()

    // 获取订阅频道
    channels, _ := store.S.Channels().List(ctx, nil)

    // 发送到所有频道
    for _, ch := range channels {
        sendNotification(ch, report)
    }

    return nil
}
```

## 安全建议

### 1. Token 保护

```yaml
# 不要在代码中硬编码 Token
# 使用环境变量或配置文件
bot:
  telegram: ${TELEGRAM_BOT_TOKEN}
  discord: ${DISCORD_BOT_TOKEN}
```

### 2. 权限最小化

- 只授予 Bot 必要的权限
- 定期审查 Bot 权限
- 使用管理员白名单

### 3. 输入验证

```go
func (ctrl *ServerController) ProcessInput(c telebot.Context) error {
    input := c.Text()

    // 验证输入
    if len(input) > 1000 {
        return c.Send("输入过长")
    }

    // 过滤特殊字符
    input = sanitize(input)

    // 处理输入
    return nil
}
```

## 与其他服务集成

### 调用 API Server

```go
import "bingo/internal/apiserver/biz"

func (ctrl *ServerController) GetUserInfo(c telebot.Context) error {
    // 通过 Store 访问数据
    user, err := store.S.Users().Get(ctx, userID)
    if err != nil {
        return c.Send("用户不存在")
    }

    return c.Send(fmt.Sprintf("用户: %s", user.Username))
}
```

### 发送邮件

```go
import "bingo/internal/pkg/mail"

func (ctrl *ServerController) SendReport(c telebot.Context) error {
    // 生成报告
    report := generateReport()

    // 发送邮件
    err := mail.Send(mail.Message{
        To:      []string{"admin@example.com"},
        Subject: "Bot 报告",
        Body:    report,
    })

    if err != nil {
        return c.Send("发送失败")
    }

    return c.Send("报告已发送")
}
```

## 相关资源

- [Telegram Bot API 文档](https://core.telegram.org/bots/api)
- [Discord Developer Portal](https://discord.com/developers/docs)
- [Telebot 库文档](https://pkg.go.dev/gopkg.in/telebot.v3)
- [DiscordGo 库文档](https://pkg.go.dev/github.com/bwmarrin/discordgo)

## 下一步

- 了解 [Scheduler 调度器](/essentials/scheduler) 如何发送定时通知
- 学习 [API Server](/essentials/apiserver) 的接口调用
