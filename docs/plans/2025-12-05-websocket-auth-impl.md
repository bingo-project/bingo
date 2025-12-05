# WebSocket 认证与状态管理实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现 WebSocket 连接后认证、状态机、Topic 订阅等功能

**Architecture:**
- 连接先进入匿名状态，10 秒内必须登录
- 登录支持密码和 Token 两种方式，Platform 从登录参数或 Token 获取
- Topic 订阅通过 channel 串行化处理，避免锁竞争

**Tech Stack:** Go, gorilla/websocket, JSON-RPC 2.0

**设计文档:** [docs/zh/advanced/websocket-heartbeat.md](../zh/advanced/websocket-heartbeat.md)

---

## Task 1: 添加 Platform 常量和验证

**Files:**
- Create: `pkg/ws/platform.go`
- Test: `pkg/ws/platform_test.go`

**Step 1: 写失败测试**

```go
// pkg/ws/platform_test.go
// ABOUTME: Tests for platform validation.
// ABOUTME: Validates platform constants and IsValidPlatform function.

package ws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidPlatform(t *testing.T) {
	tests := []struct {
		platform string
		valid    bool
	}{
		{PlatformWeb, true},
		{PlatformIOS, true},
		{PlatformAndroid, true},
		{PlatformH5, true},
		{PlatformMiniApp, true},
		{PlatformDesktop, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			assert.Equal(t, tt.valid, IsValidPlatform(tt.platform))
		})
	}
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./pkg/ws/... -run TestIsValidPlatform -v`
Expected: FAIL - undefined: PlatformWeb, IsValidPlatform

**Step 3: 实现代码**

```go
// pkg/ws/platform.go
// ABOUTME: Platform constants for client identification.
// ABOUTME: Defines valid platforms and validation function.

package ws

// Platform constants
const (
	PlatformWeb     = "web"
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
	PlatformH5      = "h5"
	PlatformMiniApp = "miniapp"
	PlatformDesktop = "desktop"
)

// IsValidPlatform checks if the platform string is valid.
func IsValidPlatform(p string) bool {
	switch p {
	case PlatformWeb, PlatformIOS, PlatformAndroid, PlatformH5, PlatformMiniApp, PlatformDesktop:
		return true
	}
	return false
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./pkg/ws/... -run TestIsValidPlatform -v`
Expected: PASS

**Step 5: 提交**

```bash
git add pkg/ws/platform.go pkg/ws/platform_test.go
git commit -m "feat(ws): add platform constants and validation"
```

---

## Task 2: 添加 HubConfig 配置结构

**Files:**
- Create: `pkg/ws/config.go`
- Test: `pkg/ws/config_test.go`

**Step 1: 写失败测试**

```go
// pkg/ws/config_test.go
// ABOUTME: Tests for Hub configuration.
// ABOUTME: Validates default config values.

package ws

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultHubConfig(t *testing.T) {
	cfg := DefaultHubConfig()

	assert.Equal(t, 10*time.Second, cfg.AnonymousTimeout)
	assert.Equal(t, 2*time.Second, cfg.AnonymousCleanup)
	assert.Equal(t, 60*time.Second, cfg.HeartbeatTimeout)
	assert.Equal(t, 30*time.Second, cfg.HeartbeatCleanup)
	assert.Equal(t, 54*time.Second, cfg.PingPeriod)
	assert.Equal(t, 60*time.Second, cfg.PongWait)
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./pkg/ws/... -run TestDefaultHubConfig -v`
Expected: FAIL - undefined: DefaultHubConfig

**Step 3: 实现代码**

```go
// pkg/ws/config.go
// ABOUTME: Configuration for WebSocket Hub.
// ABOUTME: Defines timeout and cleanup intervals.

package ws

import "time"

// HubConfig holds configuration for the Hub.
type HubConfig struct {
	// Anonymous connection timeout (must login within this time)
	AnonymousTimeout time.Duration
	// Anonymous connection cleanup interval
	AnonymousCleanup time.Duration

	// Authenticated connection heartbeat timeout
	HeartbeatTimeout time.Duration
	// Authenticated connection cleanup interval
	HeartbeatCleanup time.Duration

	// WebSocket protocol ping period
	PingPeriod time.Duration
	// WebSocket protocol pong wait timeout
	PongWait time.Duration
}

// DefaultHubConfig returns default configuration.
func DefaultHubConfig() *HubConfig {
	return &HubConfig{
		AnonymousTimeout: 10 * time.Second,
		AnonymousCleanup: 2 * time.Second,
		HeartbeatTimeout: 60 * time.Second,
		HeartbeatCleanup: 30 * time.Second,
		PingPeriod:       54 * time.Second,
		PongWait:         60 * time.Second,
	}
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./pkg/ws/... -run TestDefaultHubConfig -v`
Expected: PASS

**Step 5: 提交**

```bash
git add pkg/ws/config.go pkg/ws/config_test.go
git commit -m "feat(ws): add HubConfig with timeout settings"
```

---

## Task 3: 重构 Client 结构 - 添加 Platform 和 TokenExpiresAt

**Files:**
- Modify: `pkg/ws/client.go`
- Modify: `pkg/ws/hub_test.go`

**Step 1: 写失败测试**

```go
// 在 pkg/ws/hub_test.go 添加新测试
func TestClient_Platform(t *testing.T) {
	client := &ws.Client{
		Addr:     "127.0.0.1:8080",
		Platform: ws.PlatformIOS,
		Send:     make(chan []byte, 10),
	}

	assert.Equal(t, ws.PlatformIOS, client.Platform)
}

func TestClient_IsAuthenticated(t *testing.T) {
	client := &ws.Client{
		Addr: "127.0.0.1:8080",
		Send: make(chan []byte, 10),
	}

	// Initially not authenticated
	assert.False(t, client.IsAuthenticated())

	// After login
	client.UserID = "user-123"
	client.Platform = ws.PlatformIOS
	client.LoginTime = 1234567890
	assert.True(t, client.IsAuthenticated())
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./pkg/ws/... -run "TestClient_Platform|TestClient_IsAuthenticated" -v`
Expected: FAIL - client.Platform undefined, client.IsAuthenticated undefined

**Step 3: 修改 Client 结构**

修改 `pkg/ws/client.go`，将 `AppID uint32` 改为 `Platform string`，添加 `TokenExpiresAt int64` 和 `IsAuthenticated()` 方法：

```go
// Client represents a WebSocket client connection.
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	adapter *jsonrpc.Adapter
	ctx     context.Context

	// Send channel for outbound messages
	Send chan []byte

	// Client info
	Addr           string
	Platform       string // 替换 AppID
	UserID         string
	FirstTime      int64  // 改为 int64
	HeartbeatTime  int64  // 改为 int64
	LoginTime      int64  // 改为 int64
	TokenExpiresAt int64  // 新增

	// Subscribed topics (managed by Hub, read-only for Client)
	topics     map[string]bool
	topicsLock sync.RWMutex
}

// IsAuthenticated returns true if the client has logged in.
func (c *Client) IsAuthenticated() bool {
	return c.UserID != "" && c.Platform != "" && c.LoginTime > 0
}
```

同时更新 `NewClient`、`Login`、`Heartbeat` 等方法使用 `int64`。

**Step 4: 运行测试验证通过**

Run: `go test ./pkg/ws/... -v`
Expected: PASS

**Step 5: 提交**

```bash
git add pkg/ws/client.go pkg/ws/hub_test.go
git commit -m "refactor(ws): replace AppID with Platform, add TokenExpiresAt"
```

---

## Task 4: 重构 Hub - 添加 anonymous map 和 config

**Files:**
- Modify: `pkg/ws/hub.go`
- Modify: `pkg/ws/hub_test.go`

**Step 1: 写失败测试**

```go
// 在 pkg/ws/hub_test.go 添加
func TestHub_AnonymousCount(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHubWithConfig(ws.DefaultHubConfig())
	go hub.Run(ctx)

	client := &ws.Client{
		Addr: "127.0.0.1:8080",
		Send: make(chan []byte, 10),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	// Client is in anonymous state
	assert.Equal(t, 1, hub.AnonymousCount())
	assert.Equal(t, 0, hub.ClientCount())
}

func TestHub_AnonymousToAuthenticated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHubWithConfig(ws.DefaultHubConfig())
	go hub.Run(ctx)

	client := &ws.Client{
		Addr: "127.0.0.1:8080",
		Send: make(chan []byte, 10),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, hub.AnonymousCount())

	// Login moves client from anonymous to authenticated
	hub.Login <- &ws.LoginEvent{
		Client:   client,
		UserID:   "user-123",
		Platform: ws.PlatformIOS,
	}
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, hub.AnonymousCount())
	assert.Equal(t, 1, hub.ClientCount())
	assert.Equal(t, 1, hub.UserCount())
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./pkg/ws/... -run "TestHub_Anonymous" -v`
Expected: FAIL - undefined: NewHubWithConfig, AnonymousCount

**Step 3: 重构 Hub 结构**

```go
// Hub maintains the set of active clients and manages their lifecycle.
type Hub struct {
	config *HubConfig

	// Anonymous connections (not yet logged in)
	anonymous     map[*Client]bool
	anonymousLock sync.RWMutex

	// Authenticated connections
	clients     map[*Client]bool
	clientsLock sync.RWMutex

	// Logged-in users (key: platform_userID)
	users    map[string]*Client
	userLock sync.RWMutex

	// Topic subscriptions
	topics     map[string]map[*Client]bool
	topicsLock sync.RWMutex

	// Channels for events
	Register   chan *Client
	Unregister chan *Client
	Login      chan *LoginEvent
	Broadcast  chan []byte
}

// LoginEvent represents a user login event.
type LoginEvent struct {
	Client         *Client
	UserID         string
	Platform       string
	TokenExpiresAt int64
}

// NewHub creates a new Hub with default config.
func NewHub() *Hub {
	return NewHubWithConfig(DefaultHubConfig())
}

// NewHubWithConfig creates a new Hub with custom config.
func NewHubWithConfig(cfg *HubConfig) *Hub {
	return &Hub{
		config:     cfg,
		anonymous:  make(map[*Client]bool),
		clients:    make(map[*Client]bool),
		users:      make(map[string]*Client),
		topics:     make(map[string]map[*Client]bool),
		Register:   make(chan *Client, 256),
		Unregister: make(chan *Client, 256),
		Login:      make(chan *LoginEvent, 256),
		Broadcast:  make(chan []byte, 256),
	}
}

// AnonymousCount returns the number of anonymous connections.
func (h *Hub) AnonymousCount() int {
	h.anonymousLock.RLock()
	defer h.anonymousLock.RUnlock()
	return len(h.anonymous)
}
```

更新 `handleRegister` 将 client 放入 anonymous map，`handleLogin` 从 anonymous 移到 clients。

**Step 4: 运行测试验证通过**

Run: `go test ./pkg/ws/... -v`
Expected: PASS

**Step 5: 提交**

```bash
git add pkg/ws/hub.go pkg/ws/hub_test.go
git commit -m "refactor(ws): add anonymous map and config to Hub"
```

---

## Task 5: 实现匿名连接超时清理

**Files:**
- Modify: `pkg/ws/hub.go`
- Modify: `pkg/ws/hub_test.go`

**Step 1: 写失败测试**

```go
func TestHub_AnonymousTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use short timeout for testing
	cfg := &ws.HubConfig{
		AnonymousTimeout: 50 * time.Millisecond,
		AnonymousCleanup: 20 * time.Millisecond,
		HeartbeatTimeout: 60 * time.Second,
		HeartbeatCleanup: 30 * time.Second,
		PingPeriod:       54 * time.Second,
		PongWait:         60 * time.Second,
	}

	hub := ws.NewHubWithConfig(cfg)
	go hub.Run(ctx)

	client := &ws.Client{
		Addr:      "127.0.0.1:8080",
		Send:      make(chan []byte, 10),
		FirstTime: time.Now().Unix(),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, hub.AnonymousCount())

	// Wait for timeout + cleanup
	time.Sleep(100 * time.Millisecond)

	// Should be cleaned up
	assert.Equal(t, 0, hub.AnonymousCount())
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./pkg/ws/... -run TestHub_AnonymousTimeout -v`
Expected: FAIL - client not cleaned up (still 1 anonymous)

**Step 3: 实现清理逻辑**

在 `Hub.Run` 中启动两个 ticker：一个用于清理匿名连接，一个用于清理已认证连接。

```go
func (h *Hub) Run(ctx context.Context) {
	anonymousTicker := time.NewTicker(h.config.AnonymousCleanup)
	heartbeatTicker := time.NewTicker(h.config.HeartbeatCleanup)
	defer anonymousTicker.Stop()
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case <-anonymousTicker.C:
			h.cleanupAnonymous()

		case <-heartbeatTicker.C:
			h.cleanupInactiveClients()

		case client := <-h.Register:
			h.handleRegister(client)

		// ... 其他 case
		}
	}
}

func (h *Hub) cleanupAnonymous() {
	now := time.Now().Unix()
	timeout := int64(h.config.AnonymousTimeout.Seconds())

	h.anonymousLock.RLock()
	var inactive []*Client
	for client := range h.anonymous {
		if client.FirstTime+timeout <= now {
			inactive = append(inactive, client)
		}
	}
	h.anonymousLock.RUnlock()

	for _, client := range inactive {
		h.Unregister <- client
	}
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./pkg/ws/... -run TestHub_AnonymousTimeout -v`
Expected: PASS

**Step 5: 提交**

```bash
git add pkg/ws/hub.go pkg/ws/hub_test.go
git commit -m "feat(ws): implement anonymous connection timeout cleanup"
```

---

## Task 6: 实现单点登录踢人

**Files:**
- Modify: `pkg/ws/hub.go`
- Modify: `pkg/ws/hub_test.go`

**Step 1: 写失败测试**

```go
func TestHub_KickPreviousSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHubWithConfig(ws.DefaultHubConfig())
	go hub.Run(ctx)

	// First client logs in
	client1 := &ws.Client{
		Addr:      "client1",
		Send:      make(chan []byte, 10),
		FirstTime: time.Now().Unix(),
	}
	hub.Register <- client1
	time.Sleep(10 * time.Millisecond)

	hub.Login <- &ws.LoginEvent{
		Client:   client1,
		UserID:   "user-123",
		Platform: ws.PlatformIOS,
	}
	time.Sleep(10 * time.Millisecond)

	// Second client logs in with same user/platform
	client2 := &ws.Client{
		Addr:      "client2",
		Send:      make(chan []byte, 10),
		FirstTime: time.Now().Unix(),
	}
	hub.Register <- client2
	time.Sleep(10 * time.Millisecond)

	hub.Login <- &ws.LoginEvent{
		Client:   client2,
		UserID:   "user-123",
		Platform: ws.PlatformIOS,
	}
	time.Sleep(150 * time.Millisecond) // Wait for kick delay

	// First client should receive kick notification
	select {
	case msg := <-client1.Send:
		assert.Contains(t, string(msg), "session.kicked")
	default:
		t.Error("client1 should receive kick notification")
	}

	// Only client2 should remain
	assert.Equal(t, 1, hub.ClientCount())
	assert.Equal(t, 1, hub.UserCount())
	assert.Equal(t, client2, hub.GetUserClient(ws.PlatformIOS, "user-123"))
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./pkg/ws/... -run TestHub_KickPreviousSession -v`
Expected: FAIL - GetUserClient signature mismatch, no kick notification

**Step 3: 实现踢人逻辑**

更新 `GetUserClient` 签名和 `handleLogin`：

```go
// GetUserClient returns the client for a user.
func (h *Hub) GetUserClient(platform, userID string) *Client {
	h.userLock.RLock()
	defer h.userLock.RUnlock()
	return h.users[userKey(platform, userID)]
}

func userKey(platform, userID string) string {
	return platform + "_" + userID
}

func (h *Hub) handleLogin(event *LoginEvent) {
	client := event.Client
	key := userKey(event.Platform, event.UserID)

	// Remove from anonymous
	h.anonymousLock.Lock()
	delete(h.anonymous, client)
	h.anonymousLock.Unlock()

	// Update client info
	client.Platform = event.Platform
	client.UserID = event.UserID
	client.LoginTime = time.Now().Unix()
	client.TokenExpiresAt = event.TokenExpiresAt

	// Check for existing session
	h.userLock.Lock()
	oldClient := h.users[key]
	h.users[key] = client
	h.userLock.Unlock()

	// Add to clients
	h.clientsLock.Lock()
	h.clients[client] = true
	h.clientsLock.Unlock()

	// Kick old client if exists
	if oldClient != nil && oldClient != client {
		h.kickClient(oldClient, "您的账号已在其他设备登录")
	}
}

func (h *Hub) kickClient(client *Client, reason string) {
	// Send kick notification
	notification := jsonrpc.NewNotification("session.kicked", map[string]string{
		"reason": reason,
	})
	data, _ := json.Marshal(notification)

	select {
	case client.Send <- data:
	default:
	}

	// Kick after delay
	time.AfterFunc(100*time.Millisecond, func() {
		h.Unregister <- client
	})
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./pkg/ws/... -run TestHub_KickPreviousSession -v`
Expected: PASS

**Step 5: 提交**

```bash
git add pkg/ws/hub.go pkg/ws/hub_test.go
git commit -m "feat(ws): implement single sign-on kick mechanism"
```

---

## Task 7: 添加 Topic 订阅事件

**Files:**
- Modify: `pkg/ws/hub.go`
- Modify: `pkg/ws/hub_test.go`

**Step 1: 写失败测试**

```go
func TestHub_Subscribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHubWithConfig(ws.DefaultHubConfig())
	go hub.Run(ctx)

	client := &ws.Client{
		Addr:      "127.0.0.1:8080",
		Send:      make(chan []byte, 10),
		FirstTime: time.Now().Unix(),
	}
	hub.Register <- client
	hub.Login <- &ws.LoginEvent{
		Client:   client,
		UserID:   "user-123",
		Platform: ws.PlatformIOS,
	}
	time.Sleep(10 * time.Millisecond)

	// Subscribe to topics
	result := make(chan []string, 1)
	hub.Subscribe <- &ws.SubscribeEvent{
		Client: client,
		Topics: []string{"group:123", "room:lobby"},
		Result: result,
	}

	subscribed := <-result
	assert.ElementsMatch(t, []string{"group:123", "room:lobby"}, subscribed)

	// Verify topic count
	assert.Equal(t, 2, hub.TopicCount())
}

func TestHub_Unsubscribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHubWithConfig(ws.DefaultHubConfig())
	go hub.Run(ctx)

	client := &ws.Client{
		Addr:      "127.0.0.1:8080",
		Send:      make(chan []byte, 10),
		FirstTime: time.Now().Unix(),
	}
	hub.Register <- client
	hub.Login <- &ws.LoginEvent{
		Client:   client,
		UserID:   "user-123",
		Platform: ws.PlatformIOS,
	}
	time.Sleep(10 * time.Millisecond)

	// Subscribe
	result := make(chan []string, 1)
	hub.Subscribe <- &ws.SubscribeEvent{
		Client: client,
		Topics: []string{"group:123", "room:lobby"},
		Result: result,
	}
	<-result

	// Unsubscribe one topic
	hub.Unsubscribe <- &ws.UnsubscribeEvent{
		Client: client,
		Topics: []string{"group:123"},
	}
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, hub.TopicCount())
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./pkg/ws/... -run "TestHub_Subscribe|TestHub_Unsubscribe" -v`
Expected: FAIL - undefined: SubscribeEvent, UnsubscribeEvent, TopicCount

**Step 3: 实现订阅事件**

```go
// SubscribeEvent represents a topic subscription event.
type SubscribeEvent struct {
	Client *Client
	Topics []string
	Result chan []string
}

// UnsubscribeEvent represents a topic unsubscription event.
type UnsubscribeEvent struct {
	Client *Client
	Topics []string
}

// 在 Hub 结构添加
Subscribe   chan *SubscribeEvent
Unsubscribe chan *UnsubscribeEvent

// 在 NewHubWithConfig 中初始化
Subscribe:   make(chan *SubscribeEvent, 256),
Unsubscribe: make(chan *UnsubscribeEvent, 256),

// 在 Run 的 select 中添加
case event := <-h.Subscribe:
	subscribed := h.doSubscribe(event.Client, event.Topics)
	if event.Result != nil {
		event.Result <- subscribed
	}

case event := <-h.Unsubscribe:
	h.doUnsubscribe(event.Client, event.Topics)

// 实现 doSubscribe 和 doUnsubscribe
func (h *Hub) doSubscribe(client *Client, topics []string) []string {
	var subscribed []string
	for _, topic := range topics {
		if h.topics[topic] == nil {
			h.topics[topic] = make(map[*Client]bool)
		}
		h.topics[topic][client] = true

		client.topicsLock.Lock()
		if client.topics == nil {
			client.topics = make(map[string]bool)
		}
		client.topics[topic] = true
		client.topicsLock.Unlock()

		subscribed = append(subscribed, topic)
	}
	return subscribed
}

func (h *Hub) doUnsubscribe(client *Client, topics []string) {
	for _, topic := range topics {
		if clients, ok := h.topics[topic]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.topics, topic)
			}
		}

		client.topicsLock.Lock()
		delete(client.topics, topic)
		client.topicsLock.Unlock()
	}
}

// TopicCount returns the number of topics with subscribers.
func (h *Hub) TopicCount() int {
	return len(h.topics)
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./pkg/ws/... -run "TestHub_Subscribe|TestHub_Unsubscribe" -v`
Expected: PASS

**Step 5: 提交**

```bash
git add pkg/ws/hub.go pkg/ws/hub_test.go
git commit -m "feat(ws): add topic subscribe/unsubscribe via channel"
```

---

## Task 8: 实现 PushToTopic

**Files:**
- Modify: `pkg/ws/hub.go`
- Modify: `pkg/ws/hub_test.go`

**Step 1: 写失败测试**

```go
func TestHub_PushToTopic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHubWithConfig(ws.DefaultHubConfig())
	go hub.Run(ctx)

	// Create and login two clients
	client1 := &ws.Client{Addr: "client1", Send: make(chan []byte, 10), FirstTime: time.Now().Unix()}
	client2 := &ws.Client{Addr: "client2", Send: make(chan []byte, 10), FirstTime: time.Now().Unix()}

	hub.Register <- client1
	hub.Register <- client2
	time.Sleep(10 * time.Millisecond)

	hub.Login <- &ws.LoginEvent{Client: client1, UserID: "user1", Platform: ws.PlatformIOS}
	hub.Login <- &ws.LoginEvent{Client: client2, UserID: "user2", Platform: ws.PlatformWeb}
	time.Sleep(10 * time.Millisecond)

	// Subscribe client1 to topic
	result := make(chan []string, 1)
	hub.Subscribe <- &ws.SubscribeEvent{Client: client1, Topics: []string{"group:123"}, Result: result}
	<-result

	// Push to topic
	hub.PushToTopic("group:123", "message.new", map[string]string{"content": "hello"})
	time.Sleep(10 * time.Millisecond)

	// Only client1 should receive
	assert.Equal(t, 1, len(client1.Send))
	assert.Equal(t, 0, len(client2.Send))

	msg := <-client1.Send
	assert.Contains(t, string(msg), "message.new")
	assert.Contains(t, string(msg), "hello")
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./pkg/ws/... -run TestHub_PushToTopic -v`
Expected: FAIL - undefined: PushToTopic

**Step 3: 实现 PushToTopic**

```go
// PushToTopic sends a message to all subscribers of a topic.
func (h *Hub) PushToTopic(topic, method string, data any) {
	notification := jsonrpc.NewNotification(method, data)
	msg, err := json.Marshal(notification)
	if err != nil {
		return
	}

	h.topicsLock.RLock()
	clients := h.topics[topic]
	h.topicsLock.RUnlock()

	for client := range clients {
		select {
		case client.Send <- msg:
		default:
		}
	}
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./pkg/ws/... -run TestHub_PushToTopic -v`
Expected: PASS

**Step 5: 提交**

```bash
git add pkg/ws/hub.go pkg/ws/hub_test.go
git commit -m "feat(ws): implement PushToTopic for topic broadcasting"
```

---

## Task 9: 实现 PushToUser 和 PushToUserAllPlatforms

**Files:**
- Modify: `pkg/ws/hub.go`
- Modify: `pkg/ws/hub_test.go`

**Step 1: 写失败测试**

```go
func TestHub_PushToUser(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHubWithConfig(ws.DefaultHubConfig())
	go hub.Run(ctx)

	client := &ws.Client{Addr: "client1", Send: make(chan []byte, 10), FirstTime: time.Now().Unix()}
	hub.Register <- client
	hub.Login <- &ws.LoginEvent{Client: client, UserID: "user-123", Platform: ws.PlatformIOS}
	time.Sleep(10 * time.Millisecond)

	hub.PushToUser(ws.PlatformIOS, "user-123", "order.created", map[string]string{"order_id": "123"})
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, len(client.Send))
	msg := <-client.Send
	assert.Contains(t, string(msg), "order.created")
}

func TestHub_PushToUserAllPlatforms(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHubWithConfig(ws.DefaultHubConfig())
	go hub.Run(ctx)

	// Same user on two platforms
	client1 := &ws.Client{Addr: "client1", Send: make(chan []byte, 10), FirstTime: time.Now().Unix()}
	client2 := &ws.Client{Addr: "client2", Send: make(chan []byte, 10), FirstTime: time.Now().Unix()}

	hub.Register <- client1
	hub.Register <- client2
	time.Sleep(10 * time.Millisecond)

	hub.Login <- &ws.LoginEvent{Client: client1, UserID: "user-123", Platform: ws.PlatformIOS}
	hub.Login <- &ws.LoginEvent{Client: client2, UserID: "user-123", Platform: ws.PlatformWeb}
	time.Sleep(10 * time.Millisecond)

	hub.PushToUserAllPlatforms("user-123", "security.alert", map[string]string{"message": "new login"})
	time.Sleep(10 * time.Millisecond)

	// Both clients should receive
	assert.Equal(t, 1, len(client1.Send))
	assert.Equal(t, 1, len(client2.Send))
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./pkg/ws/... -run "TestHub_PushToUser" -v`
Expected: FAIL - undefined: PushToUser, PushToUserAllPlatforms

**Step 3: 实现推送方法**

```go
// PushToUser sends a message to a specific user on a specific platform.
func (h *Hub) PushToUser(platform, userID, method string, data any) {
	client := h.GetUserClient(platform, userID)
	if client == nil {
		return
	}

	notification := jsonrpc.NewNotification(method, data)
	msg, err := json.Marshal(notification)
	if err != nil {
		return
	}

	select {
	case client.Send <- msg:
	default:
	}
}

// PushToUserAllPlatforms sends a message to a user on all connected platforms.
func (h *Hub) PushToUserAllPlatforms(userID, method string, data any) {
	notification := jsonrpc.NewNotification(method, data)
	msg, err := json.Marshal(notification)
	if err != nil {
		return
	}

	h.userLock.RLock()
	defer h.userLock.RUnlock()

	// Find all clients for this user
	suffix := "_" + userID
	for key, client := range h.users {
		if len(key) > len(suffix) && key[len(key)-len(suffix):] == suffix {
			select {
			case client.Send <- msg:
			default:
			}
		}
	}
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./pkg/ws/... -run "TestHub_PushToUser" -v`
Expected: PASS

**Step 5: 提交**

```bash
git add pkg/ws/hub.go pkg/ws/hub_test.go
git commit -m "feat(ws): add PushToUser and PushToUserAllPlatforms"
```

---

## Task 10: 实现 Token 过期检测

**Files:**
- Modify: `pkg/ws/hub.go`
- Modify: `pkg/ws/hub_test.go`

**Step 1: 写失败测试**

```go
func TestHub_TokenExpiration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &ws.HubConfig{
		AnonymousTimeout: 10 * time.Second,
		AnonymousCleanup: 2 * time.Second,
		HeartbeatTimeout: 60 * time.Second,
		HeartbeatCleanup: 50 * time.Millisecond, // Fast for testing
		PingPeriod:       54 * time.Second,
		PongWait:         60 * time.Second,
	}

	hub := ws.NewHubWithConfig(cfg)
	go hub.Run(ctx)

	client := &ws.Client{Addr: "client1", Send: make(chan []byte, 10), FirstTime: time.Now().Unix()}
	hub.Register <- client

	// Login with token that expires immediately
	hub.Login <- &ws.LoginEvent{
		Client:         client,
		UserID:         "user-123",
		Platform:       ws.PlatformIOS,
		TokenExpiresAt: time.Now().Unix() - 1, // Already expired
	}
	time.Sleep(150 * time.Millisecond)

	// Should receive session.expired notification
	select {
	case msg := <-client.Send:
		assert.Contains(t, string(msg), "session.expired")
	default:
		t.Error("Should receive session.expired notification")
	}
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./pkg/ws/... -run TestHub_TokenExpiration -v`
Expected: FAIL - no session.expired notification

**Step 3: 在 cleanupInactiveClients 中添加 token 过期检测**

```go
func (h *Hub) cleanupInactiveClients() {
	now := time.Now().Unix()
	heartbeatTimeout := int64(h.config.HeartbeatTimeout.Seconds())

	h.clientsLock.RLock()
	var inactive []*Client
	var expired []*Client

	for client := range h.clients {
		// Check heartbeat timeout
		if client.HeartbeatTime+heartbeatTimeout <= now {
			inactive = append(inactive, client)
			continue
		}

		// Check token expiration
		if client.TokenExpiresAt > 0 && client.TokenExpiresAt <= now {
			expired = append(expired, client)
		}
	}
	h.clientsLock.RUnlock()

	// Kick inactive clients
	for _, client := range inactive {
		h.Unregister <- client
	}

	// Notify and kick expired clients
	for _, client := range expired {
		h.expireClient(client)
	}
}

func (h *Hub) expireClient(client *Client) {
	notification := jsonrpc.NewNotification("session.expired", map[string]string{
		"reason": "Token 已过期，请重新登录",
	})
	data, _ := json.Marshal(notification)

	select {
	case client.Send <- data:
	default:
	}

	time.AfterFunc(100*time.Millisecond, func() {
		h.Unregister <- client
	})
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./pkg/ws/... -run TestHub_TokenExpiration -v`
Expected: PASS

**Step 5: 提交**

```bash
git add pkg/ws/hub.go pkg/ws/hub_test.go
git commit -m "feat(ws): add token expiration detection and notification"
```

---

## Task 11: 更新 Handler 支持连接后认证

**Files:**
- Modify: `internal/apiserver/handler/ws/handler.go`
- Modify: `pkg/ws/client.go`

**Step 1: 更新 Handler**

修改 `ServeWS` 不再在连接时认证，而是创建匿名连接：

```go
// ServeWS handles WebSocket upgrade requests.
func (h *Handler) ServeWS(c *gin.Context) {
	// 1. Create base context
	ctx := context.Background()
	ctx = contextx.WithRequestID(ctx, c.GetHeader("X-Request-ID"))

	// 2. Upgrade connection (no authentication at connect time)
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 3. Create anonymous client
	client := ws.NewClient(h.hub, conn, ctx, h.adapter)

	// 4. Register with hub (as anonymous)
	h.hub.Register <- client

	// 5. Start read/write pumps
	go client.WritePump()
	go client.ReadPump()
}
```

**Step 2: 在 Client 中添加 login 方法处理**

在 `handleMessage` 中添加 login 处理：

```go
func (c *Client) handleMessage(data []byte) {
	// ... existing panic recovery ...

	var req jsonrpc.Request
	if err := json.Unmarshal(data, &req); err != nil {
		// ... error handling ...
		return
	}

	// Update heartbeat for any message
	c.Heartbeat(time.Now().Unix())

	// Handle heartbeat
	if req.Method == "heartbeat" {
		c.sendJSON(jsonrpc.NewResponse(req.ID, map[string]any{
			"status":      "ok",
			"server_time": time.Now().Unix(),
		}))
		return
	}

	// Handle login
	if req.Method == "login" {
		c.handleLogin(&req)
		return
	}

	// Require authentication for other methods
	if !c.IsAuthenticated() {
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(401, "Unauthorized", "Login required")))
		return
	}

	// Handle subscribe/unsubscribe
	if req.Method == "subscribe" {
		c.handleSubscribe(&req)
		return
	}
	if req.Method == "unsubscribe" {
		c.handleUnsubscribe(&req)
		return
	}

	// Route through adapter for business methods
	resp := c.adapter.Handle(c.ctx, &req)
	c.sendJSON(resp)
}
```

**Step 3: 实现 handleLogin**

```go
func (c *Client) handleLogin(req *jsonrpc.Request) {
	// Parse params
	var params struct {
		Type     string `json:"type"`
		Username string `json:"username"`
		Password string `json:"password"`
		Platform string `json:"platform"`
		Token    string `json:"token"`
	}

	paramsBytes, _ := json.Marshal(req.Params)
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(400, "InvalidParams", "Invalid login params")))
		return
	}

	// TODO: Implement actual authentication logic
	// For now, just validate platform and create login event

	var platform string
	var userID string
	var tokenExpiresAt int64

	switch params.Type {
	case "token":
		// Parse token to get platform and userID
		// TODO: Implement token parsing
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(501, "NotImplemented", "Token login not yet implemented")))
		return

	case "password":
		if !IsValidPlatform(params.Platform) {
			c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
				errorsx.New(400, "InvalidPlatform", "Invalid platform")))
			return
		}
		platform = params.Platform
		// TODO: Validate username/password
		userID = params.Username // Placeholder
		tokenExpiresAt = time.Now().Add(7 * 24 * time.Hour).Unix()

	default:
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(400, "InvalidLoginType", "Type must be 'token' or 'password'")))
		return
	}

	// Send login event to hub
	c.hub.Login <- &LoginEvent{
		Client:         c,
		UserID:         userID,
		Platform:       platform,
		TokenExpiresAt: tokenExpiresAt,
	}

	// Return success response
	c.sendJSON(jsonrpc.NewResponse(req.ID, map[string]any{
		"user_id":    userID,
		"platform":   platform,
		"expires_at": tokenExpiresAt,
	}))
}
```

**Step 4: 运行所有测试**

Run: `go test ./pkg/ws/... -v`
Expected: PASS

**Step 5: 提交**

```bash
git add internal/apiserver/handler/ws/handler.go pkg/ws/client.go
git commit -m "feat(ws): implement post-connect authentication flow"
```

---

## Task 12: 实现 subscribe/unsubscribe 处理

**Files:**
- Modify: `pkg/ws/client.go`

**Step 1: 实现 handleSubscribe 和 handleUnsubscribe**

```go
func (c *Client) handleSubscribe(req *jsonrpc.Request) {
	var params struct {
		Topics []string `json:"topics"`
	}

	paramsBytes, _ := json.Marshal(req.Params)
	if err := json.Unmarshal(paramsBytes, &params); err != nil || len(params.Topics) == 0 {
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(400, "InvalidParams", "topics is required")))
		return
	}

	result := make(chan []string, 1)
	c.hub.Subscribe <- &SubscribeEvent{
		Client: c,
		Topics: params.Topics,
		Result: result,
	}

	subscribed := <-result
	c.sendJSON(jsonrpc.NewResponse(req.ID, map[string]any{
		"subscribed": subscribed,
	}))
}

func (c *Client) handleUnsubscribe(req *jsonrpc.Request) {
	var params struct {
		Topics []string `json:"topics"`
	}

	paramsBytes, _ := json.Marshal(req.Params)
	if err := json.Unmarshal(paramsBytes, &params); err != nil || len(params.Topics) == 0 {
		c.sendJSON(jsonrpc.NewErrorResponse(req.ID,
			errorsx.New(400, "InvalidParams", "topics is required")))
		return
	}

	c.hub.Unsubscribe <- &UnsubscribeEvent{
		Client: c,
		Topics: params.Topics,
	}

	c.sendJSON(jsonrpc.NewResponse(req.ID, map[string]any{
		"unsubscribed": params.Topics,
	}))
}
```

**Step 2: 运行测试**

Run: `go test ./pkg/ws/... -v`
Expected: PASS

**Step 3: 提交**

```bash
git add pkg/ws/client.go
git commit -m "feat(ws): implement subscribe/unsubscribe message handlers"
```

---

## Task 13: 清理断开连接的订阅

**Files:**
- Modify: `pkg/ws/hub.go`
- Modify: `pkg/ws/hub_test.go`

**Step 1: 写失败测试**

```go
func TestHub_UnsubscribeAllOnDisconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHubWithConfig(ws.DefaultHubConfig())
	go hub.Run(ctx)

	client := &ws.Client{Addr: "client1", Send: make(chan []byte, 10), FirstTime: time.Now().Unix()}
	hub.Register <- client
	hub.Login <- &ws.LoginEvent{Client: client, UserID: "user-123", Platform: ws.PlatformIOS}
	time.Sleep(10 * time.Millisecond)

	// Subscribe to topics
	result := make(chan []string, 1)
	hub.Subscribe <- &ws.SubscribeEvent{Client: client, Topics: []string{"group:123", "room:456"}, Result: result}
	<-result
	assert.Equal(t, 2, hub.TopicCount())

	// Disconnect
	hub.Unregister <- client
	time.Sleep(10 * time.Millisecond)

	// Topics should be cleaned up
	assert.Equal(t, 0, hub.TopicCount())
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./pkg/ws/... -run TestHub_UnsubscribeAllOnDisconnect -v`
Expected: FAIL - TopicCount still 2

**Step 3: 在 handleUnregister 中清理订阅**

```go
func (h *Hub) handleUnregister(client *Client) {
	// Remove from anonymous
	h.anonymousLock.Lock()
	delete(h.anonymous, client)
	h.anonymousLock.Unlock()

	// Remove from clients
	h.clientsLock.Lock()
	if _, ok := h.clients[client]; ok {
		close(client.Send)
		delete(h.clients, client)
	}
	h.clientsLock.Unlock()

	// Remove from users
	if client.UserID != "" && client.Platform != "" {
		h.userLock.Lock()
		key := userKey(client.Platform, client.UserID)
		if c, ok := h.users[key]; ok && c == client {
			delete(h.users, key)
		}
		h.userLock.Unlock()
	}

	// Unsubscribe from all topics
	h.unsubscribeAll(client)
}

func (h *Hub) unsubscribeAll(client *Client) {
	client.topicsLock.RLock()
	topics := make([]string, 0, len(client.topics))
	for topic := range client.topics {
		topics = append(topics, topic)
	}
	client.topicsLock.RUnlock()

	if len(topics) > 0 {
		h.doUnsubscribe(client, topics)
	}
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./pkg/ws/... -run TestHub_UnsubscribeAllOnDisconnect -v`
Expected: PASS

**Step 5: 提交**

```bash
git add pkg/ws/hub.go pkg/ws/hub_test.go
git commit -m "feat(ws): cleanup topic subscriptions on disconnect"
```

---

## Task 14: 运行完整测试并更新 admserver handler

**Files:**
- Modify: `internal/admserver/handler/ws/handler.go`

**Step 1: 运行完整测试**

Run: `go test ./pkg/ws/... -v`
Expected: All PASS

**Step 2: 同步更新 admserver handler**

将 `internal/apiserver/handler/ws/handler.go` 的改动同步到 `internal/admserver/handler/ws/handler.go`。

**Step 3: 运行完整项目测试**

Run: `go test ./... -v`
Expected: PASS (可能有些测试需要调整)

**Step 4: 提交**

```bash
git add internal/admserver/handler/ws/handler.go
git commit -m "refactor(ws): sync admserver handler with apiserver"
```

---

## Task 15: 最终验证和文档更新

**Step 1: 运行所有测试**

Run: `go test ./... -v`
Expected: All PASS

**Step 2: 运行 linter**

Run: `go vet ./...`
Expected: No errors

**Step 3: 更新文件变更清单**

确认所有修改的文件与设计文档一致。

**Step 4: 提交**

```bash
git add -A
git commit -m "feat(ws): complete WebSocket auth and state management implementation"
```

---

## 总结

实现完成后的功能：

1. ✅ Platform 常量和验证
2. ✅ HubConfig 配置
3. ✅ Client 结构重构（Platform, TokenExpiresAt）
4. ✅ Hub 重构（anonymous map, config）
5. ✅ 匿名连接超时清理
6. ✅ 单点登录踢人
7. ✅ Topic 订阅事件
8. ✅ PushToTopic
9. ✅ PushToUser / PushToUserAllPlatforms
10. ✅ Token 过期检测
11. ✅ Handler 连接后认证
12. ✅ subscribe/unsubscribe 处理
13. ✅ 断开连接清理订阅
