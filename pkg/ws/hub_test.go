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
