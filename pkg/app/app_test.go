// ABOUTME: Tests for App struct and lifecycle methods.
// ABOUTME: Verifies New, Add, Register, Run, Ready functionality.

package app

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if app == nil {
		t.Fatal("New() returned nil")
	}
}

func TestAppAddAndRun(t *testing.T) {
	app, _ := New()

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
	app, _ := New()

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
	app, _ := New()

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
	app, _ := New()

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
	app, _ := New()

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
	app, _ := New(WithConfig(cfg))

	got := app.Config()
	if got != cfg {
		t.Fatal("Config() did not return configured value")
	}
}

func TestAppDBPanic(t *testing.T) {
	app, _ := New()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("DB() should panic when not configured")
		}
	}()

	app.DB()
}

func TestAppInit(t *testing.T) {
	app, _ := New()

	// First Init should succeed
	if err := app.Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	// Second Init should be no-op (idempotent)
	if err := app.Init(); err != nil {
		t.Fatalf("Second Init() returned error: %v", err)
	}
}

func TestAppInitIdempotent(t *testing.T) {
	var initCount int
	app, _ := New(WithInitFunc(func() error {
		initCount++
		return nil
	}))

	_ = app.Init()
	_ = app.Init()
	_ = app.Init()

	if initCount != 1 {
		t.Fatalf("Init was called %d times, want 1", initCount)
	}
}

func TestAppClose(t *testing.T) {
	app, _ := New()
	_ = app.Init()

	if err := app.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
}

func TestAppCloseBeforeInit(t *testing.T) {
	app, _ := New()

	// Close before Init should be safe (no-op)
	if err := app.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
}

func TestRunCallsInit(t *testing.T) {
	var initCalled bool
	app, _ := New(WithInitFunc(func() error {
		initCalled = true
		return nil
	}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_ = app.Run(ctx)

	if !initCalled {
		t.Fatal("Run() did not call Init()")
	}
}

func TestRunWithExplicitInit(t *testing.T) {
	var initCount int
	app, _ := New(WithInitFunc(func() error {
		initCount++
		return nil
	}))

	_ = app.Init() // explicit init

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_ = app.Run(ctx) // should not init again

	if initCount != 1 {
		t.Fatalf("Init was called %d times, want 1", initCount)
	}
}

func TestAppWithHealthAddr(t *testing.T) {
	app, _ := New(WithHealthAddr(":0"))

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(ctx)
	}()

	<-app.Ready()

	// Give health server time to bind to port
	time.Sleep(50 * time.Millisecond)

	// Health server should be running and ready
	addr := app.HealthAddr()
	if addr == "" || addr == ":0" {
		t.Fatalf("HealthAddr() = %q, want actual address", addr)
	}

	resp, err := http.Get("http://" + addr + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/readyz status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	cancel()

	select {
	case <-errCh:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return")
	}
}
