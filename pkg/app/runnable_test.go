// ABOUTME: Tests for core component interfaces.
// ABOUTME: Verifies Runnable, Registrar, and Named interface contracts.

package app

import (
	"context"
	"testing"
	"time"
)

// mockRunnable implements Runnable for testing
type mockRunnable struct {
	started  bool
	stopped  bool
	blockCtx context.Context
}

func (m *mockRunnable) Start(ctx context.Context) error {
	m.started = true
	<-ctx.Done()
	m.stopped = true
	return nil
}

// mockRegistrar implements Registrar for testing
type mockRegistrar struct {
	registered bool
}

func (m *mockRegistrar) Register(app *App) error {
	m.registered = true
	return nil
}

// mockNamed implements Named for testing
type mockNamed struct {
	name string
}

func (m *mockNamed) Name() string {
	return m.name
}

func (m *mockNamed) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func TestRunnableInterface(t *testing.T) {
	var r Runnable = &mockRunnable{}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		_ = r.Start(ctx)
		close(done)
	}()

	// Give goroutine time to start
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("Start did not return after context cancel")
	}
}

func TestRegistrarInterface(t *testing.T) {
	var r Registrar = &mockRegistrar{}
	err := r.Register(nil)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
}

func TestNamedInterface(t *testing.T) {
	m := &mockNamed{name: "test-server"}
	var n Named = m
	if n.Name() != "test-server" {
		t.Fatalf("Name() = %q, want %q", n.Name(), "test-server")
	}
}
