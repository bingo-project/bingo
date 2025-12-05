// ABOUTME: Tests for WebSocket client.
// ABOUTME: Validates client platform and authentication state.

package ws

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
