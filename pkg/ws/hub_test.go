// ABOUTME: Tests for WebSocket connection hub.
// ABOUTME: Validates client registration, login, and unregistration.

package ws_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/ws"
)

func TestHub_RegisterAndUnregister(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHub()
	go hub.Run(ctx)

	// Create mock client
	client := &ws.Client{
		Addr: "127.0.0.1:8080",
		Send: make(chan []byte, 10),
	}

	// Register client (goes to anonymous)
	hub.Register <- client
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, hub.AnonymousCount())

	// Unregister client
	hub.Unregister <- client
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 0, hub.AnonymousCount())
}

func TestHub_Login(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHub()
	go hub.Run(ctx)

	client := &ws.Client{
		Addr: "127.0.0.1:8080",
		Send: make(chan []byte, 10),
	}

	// Register first
	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	// Login
	hub.Login <- &ws.LoginEvent{
		Client:   client,
		UserID:   "user-123",
		Platform: ws.PlatformIOS,
	}
	time.Sleep(10 * time.Millisecond)

	// Verify user is tracked
	assert.Equal(t, 1, hub.UserCount())
	assert.NotNil(t, hub.GetUserClient(ws.PlatformIOS, "user-123"))
}

func TestHub_Broadcast(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHub()
	go hub.Run(ctx)

	client1 := &ws.Client{Addr: "client1", Send: make(chan []byte, 10)}
	client2 := &ws.Client{Addr: "client2", Send: make(chan []byte, 10)}

	hub.Register <- client1
	hub.Register <- client2
	time.Sleep(10 * time.Millisecond)

	// Login both clients to make them authenticated
	hub.Login <- &ws.LoginEvent{Client: client1, UserID: "user1", Platform: ws.PlatformIOS}
	hub.Login <- &ws.LoginEvent{Client: client2, UserID: "user2", Platform: ws.PlatformWeb}
	time.Sleep(10 * time.Millisecond)

	// Broadcast message
	hub.Broadcast <- []byte("hello")
	time.Sleep(10 * time.Millisecond)

	// Both clients should receive
	assert.Equal(t, 1, len(client1.Send))
	assert.Equal(t, 1, len(client2.Send))
}

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

	now := time.Now().Unix()
	client := &ws.Client{Addr: "client1", Send: make(chan []byte, 10), FirstTime: now, HeartbeatTime: now}
	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

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

func TestHub_UnsubscribeAllOnDisconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewHubWithConfig(ws.DefaultHubConfig())
	go hub.Run(ctx)

	now := time.Now().Unix()
	client := &ws.Client{Addr: "client1", Send: make(chan []byte, 10), FirstTime: now, HeartbeatTime: now}
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

func TestHub_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	hub := ws.NewHub()
	go hub.Run(ctx)

	client := &ws.Client{
		Addr: "127.0.0.1:8080",
		Send: make(chan []byte, 10),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, hub.AnonymousCount())

	// Cancel context to trigger shutdown
	cancel()
	time.Sleep(10 * time.Millisecond)

	// Verify client is cleaned up
	assert.Equal(t, 0, hub.AnonymousCount())

	// Verify Send channel is closed
	_, ok := <-client.Send
	assert.False(t, ok, "Send channel should be closed")
}
