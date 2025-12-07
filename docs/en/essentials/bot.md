---
title: Bot Service - Bingo Telegram/Discord Integration
description: Learn about Bingo Bot, a multi-platform bot service supporting Telegram and Discord with service management, notifications, and channel subscriptions.
---

# Bot Service

Bingo Bot is a multi-platform bot service supporting Telegram and Discord, providing service management, notification delivery, and channel subscription features.

## Core Features

- **Multi-Platform Support** - Telegram and Discord
- **Service Management** - Health checks, version queries, maintenance mode
- **Notification Subscription** - Channel subscription management and system notifications
- **Unified Interface** - Shared business logic and data storage across platforms

## Supported Platforms

### Telegram
Built on [telebot.v3](https://github.com/tucnak/telebot):
- Bot command interactions
- Channel message broadcasting
- Group management

### Discord
Built on [discordgo](https://github.com/bwmarrin/discordgo):
- Slash Commands
- Channel message broadcasting
- Server management

## Quick Start

### 1. Create Bot

#### Telegram Bot

1. Find [@BotFather](https://t.me/botfather) in Telegram
2. Send `/newbot` to create a new bot
3. Follow prompts to set bot name and username
4. Get Bot Token (keep it secure and never expose it publicly)

#### Discord Bot

1. Visit [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application"
3. Go to "Bot" tab, click "Add Bot"
4. Get Bot Token from "TOKEN" section
5. Enable necessary Intents (e.g., Message Content Intent)

### 2. Configuration

Create `bingo-bot.yaml`:

```yaml
# Bot server
server:
  name: bingo-bot
  mode: release
  addr: :18080
  timezone: Asia/Shanghai
  key: your-secret-key

# Bot configuration
bot:
  telegram: "YOUR_TELEGRAM_BOT_TOKEN"
  discord: "YOUR_DISCORD_BOT_TOKEN"

# MySQL configuration
mysql:
  host: mysql:3306
  username: root
  password: root
  database: bingo
  maxIdleConnections: 100
  maxOpenConnections: 100
  maxConnectionLifeTime: 10s
  logLevel: 4

# Redis configuration
redis:
  host: redis:6379
  password: ""
  database: 1

# JWT configuration
jwt:
  secretKey: your-jwt-secret-key
  ttl: 1440  # Token expiration (minutes)

# Logging
log:
  level: info
  days: 7
  format: console
  console: true
  maxSize: 100
  compress: true
  path: storage/log/bot.log

feature:
  profiling: true

# Mail service
mail:
  host: "smtp.example.com"
  port: 465
  username: "bot@example.com"
  password: "your-password"
  fromAddr: "noreply@example.com"
  fromName: "Bingo Bot"

# Verification code
code:
  length: 6
  ttl: 5       # Expiration (minutes)
  waiting: 1   # Resend wait time (minutes)
```

**⚠️ Security Warning:**
- Replace `YOUR_TELEGRAM_BOT_TOKEN` and `YOUR_DISCORD_BOT_TOKEN` with your actual tokens from BotFather and Discord Developer Portal
- **Never commit configuration files containing real tokens to Git repositories or share them publicly**
- Consider using environment variables or secret management services to store tokens
- If a token is leaked, regenerate it immediately on the respective platform

### 3. Start Service

```bash
# Use default configuration
./bingo-bot

# Specify configuration file
./bingo-bot -c /path/to/bingo-bot.yaml
```

### 4. Add Bot to Channel

#### Telegram

1. Add bot to group or channel
2. Grant necessary permissions (send messages, manage messages, etc.)

#### Discord

1. In Developer Portal, go to "OAuth2" > "URL Generator":
   - Scopes: `bot`, `applications.commands`
   - Select required Bot Permissions
2. Copy generated URL and open in browser
3. Select server and authorize

## Available Commands

### Service Management

| Command | Description | Example |
|---------|-------------|---------|
| `/ping` or `/pong` | Health check | `/ping` → `pong` |
| `/healthz` | Service status | `/healthz` → `ok` |
| `/version` | View version | `/version` → `v1.0.0` |
| `/maintenance` | Toggle maintenance | `/maintenance` → `Operation success` |

### Subscription Management

| Command | Description | Example |
|---------|-------------|---------|
| `/subscribe` | Subscribe to notifications | `/subscribe` → `Successfully subscribe` |
| `/unsubscribe` | Unsubscribe | `/unsubscribe` → `Successfully unsubscribe` |

## Usage Examples

### Telegram Bot

```
User: /ping
Bot:  pong

User: /subscribe
Bot:  Successfully subscribe, enjoy it!
```

### Discord Bot (Slash Commands)

```
/ping
Bot: pong

/subscribe
Bot: Successfully subscribe, enjoy it!
```

## Notification System

Bot service supports pushing system notifications to subscribed channels.

### Database Schema

Subscription data is stored in `bot_channels` table:

```sql
CREATE TABLE `bot_channels` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `source` varchar(20) NOT NULL COMMENT 'Platform: telegram, discord',
  `channel_id` varchar(100) NOT NULL COMMENT 'Channel ID',
  `author` text COMMENT 'Subscriber info (JSON)',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT 'Status: 1-enabled, 0-disabled',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_source_channel` (`source`, `channel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### Send Notifications

Trigger notifications from other services:

```go
import (
    "bingo/internal/pkg/store"
    "bingo/internal/pkg/model/bot"
)

// Get all subscribed channels
channels, err := store.S.Channels().List(ctx, &bot.ListChannelsOptions{
    Status: bot.StatusEnabled,
})

// Send message to each channel
for _, channel := range channels {
    sendMessage(channel.Source, channel.ChannelID, "System notification: Service updated")
}
```

## Architecture

### Layered Design

```
┌─────────────────────────────────────┐
│  Platform Layer (Telegram/Discord) │  ← Platform adapters
├─────────────────────────────────────┤
│     Handler Layer                   │  ← Handlers
├─────────────────────────────────────┤
│     Business Logic (Biz)            │  ← Business logic
├─────────────────────────────────────┤
│     Data Access (Store)             │  ← Data access
└─────────────────────────────────────┘
```

### Code Structure

```
internal/bot/
├── biz/               # Business logic
│   ├── bot/
│   └── syscfg/
├── telegram/         # Telegram platform
│   ├── handler/
│   ├── middleware/
│   ├── router.go
│   └── run.go
├── discord/          # Discord platform
│   ├── handler/
│   ├── middleware/
│   ├── client/
│   ├── router.go
│   └── run.go
├── app.go
└── run.go
```

## Advanced Features

### Custom Commands

#### 1. Add Command Handler

```go
// internal/bot/telegram/handler/custom.go
package handler

import "gopkg.in/telebot.v3"

type CustomHandler struct {
    // ...
}

func (h *CustomHandler) Hello(c telebot.Context) error {
    return c.Send("Hello, " + c.Sender().FirstName + "!")
}
```

#### 2. Register Route

```go
// internal/bot/telegram/router.go
func RegisterRouter(b *telebot.Bot) {
    customHandler := handler.NewCustomHandler(store.S)
    b.Handle("/hello", customHandler.Hello)
}
```

### Middleware

```go
// Logging middleware
func Logger(next telebot.HandlerFunc) telebot.HandlerFunc {
    return func(c telebot.Context) error {
        start := time.Now()
        err := next(c)
        log.Infof("Command: %s, User: %s, Duration: %v",
            c.Text(), c.Sender().Username, time.Since(start))
        return err
    }
}

// Register middleware
b.Use(Logger)
```

### Permission Control

```go
// Admin-only middleware
func AdminOnly(next telebot.HandlerFunc) telebot.HandlerFunc {
    return func(c telebot.Context) error {
        if !isAdmin(c.Sender().ID) {
            return c.Send("Admin only")
        }
        return next(c)
    }
}

// Use permission middleware
b.Handle("/maintenance", ctrl.ToggleMaintenance, AdminOnly)
```

## Operations & Monitoring

### View Logs

```bash
# Real-time logs
tail -f storage/log/bot.log

# Error logs
grep "ERROR" storage/log/bot.log
```

### Performance Profiling

If `profiling` is enabled:

```bash
# Heap memory
curl http://localhost:18080/debug/pprof/heap

# Goroutines
curl http://localhost:18080/debug/pprof/goroutine

# CPU profile (30s)
curl http://localhost:18080/debug/pprof/profile?seconds=30 > cpu.prof
```

### Common Issues

#### 1. Bot Not Responding

**Check:**
- Bot Token correctness
- Network connectivity
- Bot added to channel
- Sufficient permissions

#### 2. Subscription Failed

**Possible Causes:**
- Database connection failure
- Duplicate channel ID
- Insufficient permissions

#### 3. Message Send Failed

**Check:**
- Bot in channel
- Sufficient permissions
- Correct channel ID
- Rate limits (Telegram: 30 msg/sec, Discord: 5 req/sec)

## Best Practices

### 1. Error Handling

```go
func (h *ServerHandler) SomeCommand(c telebot.Context) error {
    result, err := h.b.DoSomething(ctx)
    if err != nil {
        log.Errorf("Operation failed: %v", err)
        return c.Send("Operation failed, please try again")
    }
    return c.Send(fmt.Sprintf("Success: %s", result))
}
```

### 2. Message Formatting

```go
// Telegram supports Markdown and HTML
func (h *ServerHandler) Status(c telebot.Context) error {
    message := `
*Service Status*
━━━━━━━━━━━━
🟢 API Server: Running
🟢 Scheduler: Running
🟡 Bot: Maintenance
━━━━━━━━━━━━
Updated: %s
    `
    return c.Send(fmt.Sprintf(message, time.Now().Format("15:04:05")),
        &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}
```

### 3. Rate Limiting

```go
// Prevent abuse
var limiter = rate.NewLimiter(rate.Every(time.Second), 10)

func (h *ServerHandler) RateLimitedCommand(c telebot.Context) error {
    if !limiter.Allow() {
        return c.Send("Too many requests, please try again later")
    }
    // Process command
    return nil
}
```

### 4. Scheduler Integration

Send scheduled notifications:

```go
// Define task in Scheduler
func SendDailyReport(ctx context.Context, t *asynq.Task) error {
    report := generateReport()
    channels, _ := store.S.Channels().List(ctx, nil)
    for _, ch := range channels {
        sendNotification(ch, report)
    }
    return nil
}
```

## Security

### 1. Token Protection

```yaml
# Don't hardcode tokens
# Use environment variables or config files
bot:
  telegram: ${TELEGRAM_BOT_TOKEN}
  discord: ${DISCORD_BOT_TOKEN}
```

### 2. Minimal Permissions

- Grant only necessary permissions
- Regular permission audits
- Use admin whitelist

### 3. Input Validation

```go
func (h *ServerHandler) ProcessInput(c telebot.Context) error {
    input := c.Text()
    if len(input) > 1000 {
        return c.Send("Input too long")
    }
    input = sanitize(input)
    // Process input
    return nil
}
```

## Integration with Other Services

### Call API Server

```go
import "bingo/internal/apiserver/biz"

func (h *ServerHandler) GetUserInfo(c telebot.Context) error {
    user, err := store.S.Users().Get(ctx, userID)
    if err != nil {
        return c.Send("User not found")
    }
    return c.Send(fmt.Sprintf("User: %s", user.Username))
}
```

### Send Email

```go
import "bingo/internal/pkg/mail"

func (h *ServerHandler) SendReport(c telebot.Context) error {
    report := generateReport()
    err := mail.Send(mail.Message{
        To:      []string{"admin@example.com"},
        Subject: "Bot Report",
        Body:    report,
    })
    if err != nil {
        return c.Send("Failed to send")
    }
    return c.Send("Report sent")
}
```

## Related Resources

- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Discord Developer Portal](https://discord.com/developers/docs)
- [Telebot Documentation](https://pkg.go.dev/gopkg.in/telebot.v3)
- [DiscordGo Documentation](https://pkg.go.dev/github.com/bwmarrin/discordgo)

## Next Step

- [Core Components Overview](../components/overview.md) - Learn about the framework's core components
