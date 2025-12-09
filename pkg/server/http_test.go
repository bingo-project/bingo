// ABOUTME: Tests for HTTPServer implementation.
// ABOUTME: Verifies HTTP server start, stop, and Runnable interface.

package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHTTPServerStartStop(t *testing.T) {
	engine := gin.New()
	engine.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	srv := NewHTTPServer(":0", engine) // :0 = random port

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	// Wait for server to start
	time.Sleep(50 * time.Millisecond)

	// Verify server is running
	addr := srv.Addr()
	resp, err := http.Get("http://" + addr + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /health status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

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

func TestHTTPServerName(t *testing.T) {
	srv := NewHTTPServer(":8080", nil)
	if srv.Name() != "http" {
		t.Fatalf("Name() = %q, want %q", srv.Name(), "http")
	}
}
