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

	// Register client
	hub.Register <- client
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, hub.ClientCount())

	// Unregister client
	hub.Unregister <- client
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 0, hub.ClientCount())
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

	// Broadcast message
	hub.Broadcast <- []byte("hello")
	time.Sleep(10 * time.Millisecond)

	// Both clients should receive
	assert.Equal(t, 1, len(client1.Send))
	assert.Equal(t, 1, len(client2.Send))
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
	assert.Equal(t, 1, hub.ClientCount())

	// Cancel context to trigger shutdown
	cancel()
	time.Sleep(10 * time.Millisecond)

	// Verify client is cleaned up
	assert.Equal(t, 0, hub.ClientCount())

	// Verify Send channel is closed
	_, ok := <-client.Send
	assert.False(t, ok, "Send channel should be closed")
}
