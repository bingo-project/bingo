// ABOUTME: Tests for WebSocketServer implementation.
// ABOUTME: Verifies WebSocket server start, stop, and Runnable interface.

package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestWebSocketServerStartStop(t *testing.T) {
	engine := gin.New()
	engine.GET("/ws", func(c *gin.Context) {
		c.String(http.StatusOK, "ws endpoint")
	})

	srv := NewWebSocketServer(":0", engine)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Verify server is running
	addr := srv.Addr()
	resp, err := http.Get("http://" + addr + "/ws")
	if err != nil {
		t.Fatalf("GET /ws failed: %v", err)
	}
	resp.Body.Close()

	cancel()

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after cancel")
	}
}

func TestWebSocketServerName(t *testing.T) {
	srv := NewWebSocketServer(":8080", nil)
	if srv.Name() != "websocket" {
		t.Fatalf("Name() = %q, want %q", srv.Name(), "websocket")
	}
}
