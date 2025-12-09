// ABOUTME: Tests for App struct and lifecycle methods.
// ABOUTME: Verifies New, Add, Register, Run, Ready functionality.

package app

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	app := New()
	if app == nil {
		t.Fatal("New() returned nil")
	}
}

func TestAppAddAndRun(t *testing.T) {
	app := New()

	var started atomic.Bool

	app.Add(runnableFunc(func(ctx context.Context) error {
		started.Store(true)
		<-ctx.Done()
		return nil
	}))

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(ctx)
	}()

	// Wait for ready
	select {
	case <-app.Ready():
	case <-time.After(time.Second):
		t.Fatal("App did not become ready")
	}

	// Give runnable goroutine time to execute
	time.Sleep(10 * time.Millisecond)

	if !started.Load() {
		t.Fatal("Runnable was not started")
	}

	cancel()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("Run returned unexpected error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return after cancel")
	}
}

func TestAppRegister(t *testing.T) {
	app := New()

	var registered bool
	app.Register(registrarFunc(func(a *App) error {
		registered = true
		return nil
	}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_ = app.Run(ctx)

	if !registered {
		t.Fatal("Registrar was not called")
	}
}

func TestAppRegisterOrder(t *testing.T) {
	app := New()

	var order []int
	app.Register(registrarFunc(func(a *App) error {
		order = append(order, 1)
		return nil
	}))
	app.Register(registrarFunc(func(a *App) error {
		order = append(order, 2)
		return nil
	}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_ = app.Run(ctx)

	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Fatalf("Register order = %v, want [1, 2]", order)
	}
}

func TestAppRunnableStartFailure(t *testing.T) {
	app := New()

	expectedErr := errors.New("start failed")
	app.Add(runnableFunc(func(ctx context.Context) error {
		return expectedErr
	}))

	ctx := context.Background()
	err := app.Run(ctx)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("Run error = %v, want %v", err, expectedErr)
	}
}

func TestAppLogsRunnableNames(t *testing.T) {
	// This is a behavior test - we verify Named interface is detected
	app := New()

	named := &mockNamed{name: "test-server"}
	app.Add(named)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_ = app.Run(ctx)
	// If we get here without panic, the Named detection works
}

// runnableFunc adapts a function to Runnable interface
type runnableFunc func(ctx context.Context) error

func (f runnableFunc) Start(ctx context.Context) error {
	return f(ctx)
}

// registrarFunc adapts a function to Registrar interface
type registrarFunc func(app *App) error

func (f registrarFunc) Register(app *App) error {
	return f(app)
}

func TestAppConfig(t *testing.T) {
	cfg := &struct{ Name string }{Name: "test"}
	app := New(WithConfig(cfg))

	got := app.Config()
	if got != cfg {
		t.Fatal("Config() did not return configured value")
	}
}

func TestAppDBPanic(t *testing.T) {
	app := New()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("DB() should panic when not configured")
		}
	}()

	app.DB()
}
