// ABOUTME: Tests for HealthServer implementation.
// ABOUTME: Verifies health check endpoints and readiness state.

package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestHealthServerEndpoints(t *testing.T) {
	srv := NewHealthServer(":0")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	addr := srv.Addr()

	// Test /healthz - always 200
	resp, err := http.Get("http://" + addr + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/healthz status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// Test /readyz - 503 before ready, 200 after
	resp, err = http.Get("http://" + addr + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz failed: %v", err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("/readyz status = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
	resp.Body.Close()

	// Mark ready
	srv.SetReady(true)

	resp, err = http.Get("http://" + addr + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/readyz status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if result["status"] != "ok" {
		t.Fatalf("status = %q, want %q", result["status"], "ok")
	}
}

func TestHealthServerShutdown(t *testing.T) {
	srv := NewHealthServer(":0")
	srv.SetReady(true)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	addr := srv.Addr()

	// Signal shutdown
	srv.SetReady(false)

	resp, _ := http.Get("http://" + addr + "/readyz")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("/readyz status = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var result map[string]string
	json.Unmarshal(body, &result)
	if result["status"] != "shutting_down" {
		t.Fatalf("status = %q, want %q", result["status"], "shutting_down")
	}

	cancel()
}

func TestHealthServerName(t *testing.T) {
	srv := NewHealthServer(":8081")
	if srv.Name() != "health" {
		t.Fatalf("Name() = %q, want %q", srv.Name(), "health")
	}
}

func TestHealthServerAsRunnable(t *testing.T) {
	// Verify HealthServer implements the Runnable pattern correctly
	// by testing it can be started and stopped via context
	srv := NewHealthServer(":0")

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Server should be running
	addr := srv.Addr()
	resp, err := http.Get("http://" + addr + "/healthz")
	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	resp.Body.Close()

	// Cancel context to stop server
	cancel()

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop")
	}
}
