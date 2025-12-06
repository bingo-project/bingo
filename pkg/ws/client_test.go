// ABOUTME: Tests for WebSocket client.
// ABOUTME: Validates client platform and authentication state.

package ws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
)

func TestClient_Platform(t *testing.T) {
	client := &Client{
		Addr:     "127.0.0.1:8080",
		Platform: PlatformIOS,
		Send:     make(chan []byte, 10),
	}

	assert.Equal(t, PlatformIOS, client.Platform)
}

func TestClient_IsAuthenticated(t *testing.T) {
	client := &Client{
		Addr: "127.0.0.1:8080",
		Send: make(chan []byte, 10),
	}

	// Initially not authenticated
	assert.False(t, client.IsAuthenticated())

	// After login
	client.UserID = "user-123"
	client.Platform = PlatformIOS
	client.LoginTime = 1234567890
	assert.True(t, client.IsAuthenticated())
}

func TestClient_HasID(t *testing.T) {
	hub := NewHub()

	client := NewClient(hub, nil, context.Background())

	assert.NotEmpty(t, client.ID)
	assert.Len(t, client.ID, 36) // UUID length
}

func TestClient_IDIsUnique(t *testing.T) {
	hub := NewHub()

	client1 := NewClient(hub, nil, context.Background())
	client2 := NewClient(hub, nil, context.Background())

	assert.NotEqual(t, client1.ID, client2.ID)
}

func TestClient_WithRouter(t *testing.T) {
	hub := NewHub()
	router := NewRouter()

	router.Handle("test.method", func(c *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	})

	client := NewClient(hub, nil, context.Background(), WithRouter(router))

	assert.NotNil(t, client.router)
	assert.Equal(t, router, client.router)
}
