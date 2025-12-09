# Phase 2: App Layer Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement App layer as unified entry point with Runnable/Registrar pattern, replacing current Server/Runner pattern.

**Architecture:** App coordinates component lifecycle through Runnable (long-running) and Registrar (setup-only) interfaces. Existing Server implementations become Runnable adapters. Bootstrap logic integrates into App.Init().

**Tech Stack:** Go 1.21+, errgroup, context, existing dependencies (gorm, redis, gin, grpc)

---

## 迁移策略

**渐进式迁移**：新建框架代码，不修改现有实现。

| 包 | 动作 | 说明 |
|-----|------|------|
| `pkg/app` | 新建 | App, Runnable, Registrar 接口 |
| `pkg/server` | 新建 | HTTPServer, GRPCServer 等（实现 Runnable） |
| `internal/pkg/server` | 保持不动 | 现有实现继续工作 |
| `internal/apiserver` | 保持不动 | 等框架稳定后再迁移 |

**为什么这样做**：
- 零风险：现有代码不受影响
- 独立验证：新框架可以单独测试
- 渐进迁移：框架稳定后再逐步切换

---

## Overview

Phase 2 is divided into sub-phases to enable incremental delivery:

| Sub-Phase | Description | Deliverable |
|-----------|-------------|-------------|
| 2.1 | Core interfaces (Runnable, Registrar, App) | Interfaces + basic App skeleton |
| 2.2 | Server adapters | HTTPServer, GRPCServer as Runnable |
| 2.3 | Health check | Independent health server |
| 2.4 | Signals package | SetupSignalHandler() |
| 2.5 | Integration | Migrate apiserver/admserver to App |
| 2.6 | Cleanup | Remove old Server/Runner |

## File Structure

```
pkg/
├── app/
│   ├── app.go           # App struct and lifecycle
│   ├── options.go       # WithXxx options
│   ├── runnable.go      # Runnable, Registrar, Named interfaces
│   └── signals.go       # SetupSignalHandler
└── server/
    ├── http.go          # HTTPServer (Runnable adapter)
    ├── grpc.go          # GRPCServer (Runnable adapter)
    ├── websocket.go     # WebSocketServer (Runnable adapter)
    └── health.go        # HealthServer (internal)
```

---

## Sub-Phase 2.1: Core Interfaces

### Task 1: Create Runnable and Registrar interfaces

**Files:**
- Create: `pkg/app/runnable.go`
- Test: `pkg/app/runnable_test.go`

**Step 1: Write the test**

```go
// pkg/app/runnable_test.go
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
```

**Step 2: Run test to verify it fails**

Run: `go test -v bingo/pkg/app -run TestRunnable`
Expected: FAIL - package does not exist

**Step 3: Create package with interfaces**

```go
// pkg/app/runnable.go
// ABOUTME: Core component interfaces for the App lifecycle.
// ABOUTME: Runnable for long-running components, Registrar for setup-only components.

package app

import "context"

// Runnable is a component that runs for the lifetime of the App.
// Start blocks until ctx is canceled or an error occurs.
type Runnable interface {
	Start(ctx context.Context) error
}

// Registrar is a component that registers with the App during startup.
// Register is called once before any Runnable starts.
type Registrar interface {
	Register(app *App) error
}

// Named is an optional interface for components that have a name.
// Used for logging.
type Named interface {
	Name() string
}
```

**Step 4: Run test to verify compilation**

Run: `go test -v bingo/pkg/app -run TestRunnable`
Expected: FAIL - App type not defined (this is expected, we create it next)

**Step 5: Create minimal App type stub**

```go
// pkg/app/app.go
// ABOUTME: App is the unified entry point for bingo applications.
// ABOUTME: Coordinates Runnable lifecycle and provides dependency access.

package app

// App is the main application container.
type App struct {
	// placeholder - will be filled in Task 2
}
```

**Step 6: Run test to verify it passes**

Run: `go test -v bingo/pkg/app -run Test`
Expected: PASS

**Step 7: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): add Runnable, Registrar, Named interfaces"
```

---

### Task 2: Implement App struct with lifecycle methods

**Files:**
- Modify: `pkg/app/app.go`
- Create: `pkg/app/options.go`
- Test: `pkg/app/app_test.go`

**Step 1: Write the test**

```go
// pkg/app/app_test.go
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
	r := &mockRunnable{}

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
```

**Step 2: Run test to verify it fails**

Run: `go test -v bingo/pkg/app -run TestNew`
Expected: FAIL - New() not defined

**Step 3: Implement App struct**

```go
// pkg/app/app.go
// ABOUTME: App is the unified entry point for bingo applications.
// ABOUTME: Coordinates Runnable lifecycle and provides dependency access.

package app

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

// App is the main application container.
type App struct {
	runnables  []Runnable
	registrars []Registrar

	ready     chan struct{}
	readyOnce sync.Once

	shutdownTimeout time.Duration

	mu sync.Mutex
}

// New creates a new App with the given options.
func New(opts ...Option) *App {
	app := &App{
		ready:           make(chan struct{}),
		shutdownTimeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// Add adds a Runnable to the App.
// If the Runnable also implements Registrar, it will be registered.
// Returns the App for chaining.
func (app *App) Add(r Runnable) *App {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.runnables = append(app.runnables, r)

	// If it also implements Registrar, add to registrars
	if reg, ok := r.(Registrar); ok {
		app.registrars = append(app.registrars, reg)
	}

	return app
}

// Register adds a Registrar to the App.
// Returns the App for chaining.
func (app *App) Register(r Registrar) *App {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.registrars = append(app.registrars, r)
	return app
}

// Run starts all runnables and blocks until ctx is canceled.
func (app *App) Run(ctx context.Context) error {
	// Phase 1: Register (serial, in order)
	for _, reg := range app.registrars {
		if err := reg.Register(app); err != nil {
			return err
		}
	}

	// Phase 2: Start all runnables (concurrent)
	if len(app.runnables) == 0 {
		app.markReady()
		<-ctx.Done()
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	for _, r := range app.runnables {
		r := r
		g.Go(func() error {
			return r.Start(ctx)
		})
	}

	// Mark ready after all runnables started
	app.markReady()

	return g.Wait()
}

// Ready returns a channel that is closed when the App is ready.
func (app *App) Ready() <-chan struct{} {
	return app.ready
}

func (app *App) markReady() {
	app.readyOnce.Do(func() {
		close(app.ready)
	})
}
```

**Step 4: Create options file**

```go
// pkg/app/options.go
// ABOUTME: Option functions for configuring App.
// ABOUTME: Use WithXxx pattern for optional configuration.

package app

import "time"

// Option configures an App.
type Option func(*App)

// WithShutdownTimeout sets the shutdown timeout.
func WithShutdownTimeout(d time.Duration) Option {
	return func(app *App) {
		app.shutdownTimeout = d
	}
}
```

**Step 5: Add time import to app.go**

Add `"time"` to imports in app.go.

**Step 6: Run test to verify it passes**

Run: `go test -v bingo/pkg/app`
Expected: PASS

**Step 7: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): implement App with Add, Register, Run, Ready"
```

---

### Task 3: Add logging support to App

**Files:**
- Modify: `pkg/app/app.go`
- Modify: `pkg/app/options.go`
- Modify: `pkg/app/app_test.go`

**Step 1: Write the test**

Add to `pkg/app/app_test.go`:

```go
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
```

**Step 2: Run test**

Run: `go test -v bingo/pkg/app -run TestAppLogsRunnableNames`
Expected: PASS (current impl doesn't log yet, but should pass)

**Step 3: Add getName helper**

Add to `pkg/app/app.go`:

```go
// getName returns the name of a component if it implements Named.
func getName(r any) string {
	if n, ok := r.(Named); ok {
		return n.Name()
	}
	return ""
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v bingo/pkg/app`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): add getName helper for Named interface"
```

---

## Sub-Phase 2.2: Server Adapters

### Task 4: Create HTTPServer as Runnable

**Files:**
- Create: `pkg/server/http.go`
- Test: `pkg/server/http_test.go`

**Step 1: Write the test**

```go
// pkg/server/http_test.go
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
```

**Step 2: Run test to verify it fails**

Run: `go test -v bingo/pkg/server -run TestHTTPServer`
Expected: FAIL - package does not exist

**Step 3: Implement HTTPServer**

```go
// pkg/server/http.go
// ABOUTME: HTTP server implementation as Runnable.
// ABOUTME: Wraps gin.Engine with graceful shutdown support.

package server

import (
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTPServer is an HTTP server that implements Runnable.
type HTTPServer struct {
	addr     string
	engine   *gin.Engine
	server   *http.Server
	listener net.Listener
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(addr string, engine *gin.Engine) *HTTPServer {
	return &HTTPServer{
		addr:   addr,
		engine: engine,
	}
}

// Start starts the HTTP server and blocks until ctx is canceled.
func (s *HTTPServer) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	s.server = &http.Server{
		Handler: s.engine,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.Serve(ln)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return s.server.Shutdown(context.Background())
	}
}

// Name returns the server name for logging.
func (s *HTTPServer) Name() string {
	return "http"
}

// Addr returns the actual listen address.
// Only valid after Start is called.
func (s *HTTPServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v bingo/pkg/server -run TestHTTPServer`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/server/
git commit -m "feat(server): add HTTPServer as Runnable"
```

---

### Task 5: Create GRPCServer as Runnable

**Files:**
- Create: `pkg/server/grpc.go`
- Test: `pkg/server/grpc_test.go`

**Step 1: Write the test**

```go
// pkg/server/grpc_test.go
package server

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPCServerStartStop(t *testing.T) {
	srv := NewGRPCServer(":0", grpc.NewServer())

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Verify server is running by attempting connection
	addr := srv.Addr()
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.Dial failed: %v", err)
	}
	conn.Close()

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after cancel")
	}
}

func TestGRPCServerName(t *testing.T) {
	srv := NewGRPCServer(":9090", nil)
	if srv.Name() != "grpc" {
		t.Fatalf("Name() = %q, want %q", srv.Name(), "grpc")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v bingo/pkg/server -run TestGRPCServer`
Expected: FAIL - GRPCServer not defined

**Step 3: Implement GRPCServer**

```go
// pkg/server/grpc.go
// ABOUTME: gRPC server implementation as Runnable.
// ABOUTME: Wraps grpc.Server with graceful shutdown support.

package server

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

// GRPCServer is a gRPC server that implements Runnable.
type GRPCServer struct {
	addr     string
	server   *grpc.Server
	listener net.Listener
}

// NewGRPCServer creates a new gRPC server.
func NewGRPCServer(addr string, server *grpc.Server) *GRPCServer {
	return &GRPCServer{
		addr:   addr,
		server: server,
	}
}

// Start starts the gRPC server and blocks until ctx is canceled.
func (s *GRPCServer) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.Serve(ln)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.server.GracefulStop()
		return nil
	}
}

// Name returns the server name for logging.
func (s *GRPCServer) Name() string {
	return "grpc"
}

// Addr returns the actual listen address.
func (s *GRPCServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v bingo/pkg/server -run TestGRPCServer`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/server/
git commit -m "feat(server): add GRPCServer as Runnable"
```

---

### Task 6: Create WebSocketServer as Runnable

**Files:**
- Create: `pkg/server/websocket.go`
- Test: `pkg/server/websocket_test.go`

**Step 1: Write the test**

```go
// pkg/server/websocket_test.go
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
```

**Step 2: Run test to verify it fails**

Run: `go test -v bingo/pkg/server -run TestWebSocketServer`
Expected: FAIL - WebSocketServer not defined

**Step 3: Implement WebSocketServer**

```go
// pkg/server/websocket.go
// ABOUTME: WebSocket server implementation as Runnable.
// ABOUTME: Wraps gin.Engine for WebSocket connections.

package server

import (
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebSocketServer is a WebSocket server that implements Runnable.
type WebSocketServer struct {
	addr     string
	engine   *gin.Engine
	server   *http.Server
	listener net.Listener
}

// NewWebSocketServer creates a new WebSocket server.
func NewWebSocketServer(addr string, engine *gin.Engine) *WebSocketServer {
	return &WebSocketServer{
		addr:   addr,
		engine: engine,
	}
}

// Start starts the WebSocket server and blocks until ctx is canceled.
func (s *WebSocketServer) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	s.server = &http.Server{
		Handler: s.engine,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.Serve(ln)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return s.server.Shutdown(context.Background())
	}
}

// Name returns the server name for logging.
func (s *WebSocketServer) Name() string {
	return "websocket"
}

// Addr returns the actual listen address.
func (s *WebSocketServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v bingo/pkg/server -run TestWebSocketServer`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/server/
git commit -m "feat(server): add WebSocketServer as Runnable"
```

---

## Sub-Phase 2.3: Health Check Server

### Task 7: Create HealthServer

**Files:**
- Create: `pkg/server/health.go`
- Test: `pkg/server/health_test.go`

**Step 1: Write the test**

```go
// pkg/server/health_test.go
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
```

**Step 2: Run test to verify it fails**

Run: `go test -v bingo/pkg/server -run TestHealthServer`
Expected: FAIL - HealthServer not defined

**Step 3: Implement HealthServer**

```go
// pkg/server/health.go
// ABOUTME: Health check server for K8s liveness and readiness probes.
// ABOUTME: Runs on independent port, provides /healthz and /readyz endpoints.

package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sync/atomic"
)

// HealthServer provides health check endpoints.
type HealthServer struct {
	addr     string
	server   *http.Server
	listener net.Listener
	ready    atomic.Bool
}

// NewHealthServer creates a new health check server.
func NewHealthServer(addr string) *HealthServer {
	return &HealthServer{
		addr: addr,
	}
}

// SetReady sets the readiness state.
func (s *HealthServer) SetReady(ready bool) {
	s.ready.Store(ready)
}

// Start starts the health server and blocks until ctx is canceled.
func (s *HealthServer) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)

	s.server = &http.Server{
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.Serve(ln)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return s.server.Shutdown(context.Background())
	}
}

// Name returns the server name for logging.
func (s *HealthServer) Name() string {
	return "health"
}

// Addr returns the actual listen address.
func (s *HealthServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

func (s *HealthServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *HealthServer) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.ready.Load() {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "shutting_down"})
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v bingo/pkg/server -run TestHealthServer`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/server/
git commit -m "feat(server): add HealthServer for K8s probes"
```

---

## Sub-Phase 2.4: Signals Package

### Task 8: Create SetupSignalHandler

**Files:**
- Create: `pkg/app/signals.go`
- Test: `pkg/app/signals_test.go`

**Step 1: Write the test**

```go
// pkg/app/signals_test.go
package app

import (
	"syscall"
	"testing"
	"time"
)

func TestSetupSignalHandler(t *testing.T) {
	ctx := SetupSignalHandler()

	// Context should not be done initially
	select {
	case <-ctx.Done():
		t.Fatal("context should not be done initially")
	default:
		// expected
	}

	// Send SIGINT
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	// Context should be done after signal
	select {
	case <-ctx.Done():
		// expected
	case <-time.After(time.Second):
		t.Fatal("context should be done after SIGINT")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v bingo/pkg/app -run TestSetupSignalHandler`
Expected: FAIL - SetupSignalHandler not defined

**Step 3: Implement SetupSignalHandler**

```go
// pkg/app/signals.go
// ABOUTME: Signal handling utilities for graceful shutdown.
// ABOUTME: SetupSignalHandler returns context that cancels on SIGTERM/SIGINT.

package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler returns a context that is canceled when SIGTERM or SIGINT is received.
// On second signal, os.Exit(1) is called immediately.
func SetupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c
		cancel() // First signal: graceful shutdown
		<-c
		os.Exit(1) // Second signal: force exit
	}()

	return ctx
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v bingo/pkg/app -run TestSetupSignalHandler`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): add SetupSignalHandler for graceful shutdown"
```

---

## Sub-Phase 2.5: Integration

### Task 9: Add dependency access methods to App

**Files:**
- Modify: `pkg/app/app.go`
- Modify: `pkg/app/options.go`
- Test: `pkg/app/app_test.go`

**Step 1: Write the test**

Add to `pkg/app/app_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test -v bingo/pkg/app -run TestAppConfig`
Expected: FAIL - WithConfig not defined

**Step 3: Add dependency fields and methods**

Update `pkg/app/app.go`:

```go
// Add to App struct:
	config any
	db     *gorm.DB
	cache  any

// Add methods:

// Config returns the application configuration.
func (app *App) Config() any {
	return app.config
}

// DB returns the database connection.
// Panics if not configured.
func (app *App) DB() *gorm.DB {
	if app.db == nil {
		panic("database not configured")
	}
	return app.db
}
```

Update `pkg/app/options.go`:

```go
import "gorm.io/gorm"

// WithConfig sets the application configuration.
func WithConfig(cfg any) Option {
	return func(app *App) {
		app.config = cfg
	}
}

// WithDB sets the database connection.
func WithDB(db *gorm.DB) Option {
	return func(app *App) {
		app.db = db
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v bingo/pkg/app -run TestAppConfig`
Run: `go test -v bingo/pkg/app -run TestAppDBPanic`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): add Config and DB dependency access methods"
```

---

### Task 10: Integrate HealthServer into App

**Files:**
- Modify: `pkg/app/app.go`
- Modify: `pkg/app/options.go`
- Test: `pkg/app/app_test.go`

**Step 1: Write the test**

Add to `pkg/app/app_test.go`:

```go
func TestAppHealthServer(t *testing.T) {
	app := New(WithHealthAddr(":0"))

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(ctx)
	}()

	<-app.Ready()

	// Health server should be running
	// Note: we'd need to expose the addr to test properly
	// For now, just verify no panic

	cancel()

	select {
	case <-errCh:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return")
	}
}
```

**Step 2: Add health server integration**

Update `pkg/app/app.go`:

```go
// Add to App struct:
	healthAddr   string
	healthServer *server.HealthServer

// Update Run method to start health server first and set ready:

func (app *App) Run(ctx context.Context) error {
	// Start health server first (if configured)
	if app.healthAddr != "" {
		app.healthServer = server.NewHealthServer(app.healthAddr)
		go app.healthServer.Start(ctx)
	}

	// ... existing Register phase ...

	// ... existing Start phase ...

	// Mark ready (updates health server too)
	app.markReady()

	// ... rest of method ...
}

func (app *App) markReady() {
	app.readyOnce.Do(func() {
		if app.healthServer != nil {
			app.healthServer.SetReady(true)
		}
		close(app.ready)
	})
}
```

Update `pkg/app/options.go`:

```go
// WithHealthAddr enables the health server on the given address.
func WithHealthAddr(addr string) Option {
	return func(app *App) {
		app.healthAddr = addr
	}
}
```

**Step 3: Run test to verify it passes**

Run: `go test -v bingo/pkg/app -run TestAppHealthServer`
Expected: PASS

**Step 4: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): integrate HealthServer into App lifecycle"
```

---

## Sub-Phase 2.6: Migration (Integration with existing code)

### Task 11: Create migration example for apiserver

This task demonstrates how to migrate apiserver to use the new App pattern. The actual migration should be done incrementally.

**Files:**
- Document migration pattern (no code changes yet)

**Migration Pattern:**

Before (current):
```go
func run() error {
    ctx, stop := signal.NotifyContext(...)
    defer stop()

    ginEngine := initGinEngine()
    grpcServer := initGRPCServer(...)
    wsEngine, wsHub := initWebSocket()

    runner := server.Assemble(...)
    return runner.Run(ctx)
}
```

After (new pattern):
```go
func run() error {
    app := bingo.New(
        bingo.WithConfig(&facade.Config),
        bingo.WithHealthAddr(":8081"),
    )

    // Add servers
    app.Add(server.NewHTTPServer(facade.Config.HTTP.Addr, initGinEngine()))
    app.Add(server.NewGRPCServer(facade.Config.GRPC.Addr, initGRPCServer()))
    app.Add(server.NewWebSocketServer(facade.Config.WebSocket.Addr, initWebSocket()))

    return app.Run(bingo.SetupSignalHandler())
}
```

**Commit:**

```bash
git commit --allow-empty -m "docs: document migration pattern for apiserver"
```

---

## Summary

This plan covers the core Phase 2 implementation:

1. **Core Interfaces** (Tasks 1-3): Runnable, Registrar, Named, App basics
2. **Server Adapters** (Tasks 4-6): HTTP, gRPC, WebSocket servers as Runnable
3. **Health Check** (Task 7): Independent health server
4. **Signals** (Task 8): SetupSignalHandler
5. **Integration** (Tasks 9-10): Dependency access, health integration
6. **Migration** (Task 11): Documentation for migrating existing code

Total: ~11 tasks, each with TDD approach (test first, implement, commit).

## Notes

- This plan creates new `pkg/app` and `pkg/server` packages
- Existing `internal/pkg/server` remains unchanged during implementation
- Migration happens after new packages are stable
- Phase 3 (separating starter) happens after Phase 2 is complete

---

## Sub-Phase 2.7: Complete App Lifecycle (补充)

根据设计文档 review，补充以下缺失的 tasks。

### Task 12: Update New() to return error

**Files:**
- Modify: `pkg/app/app.go`
- Modify: `pkg/app/app_test.go`

**Step 1: Update tests**

```go
// Update existing tests to handle error return
func TestNew(t *testing.T) {
    app, err := New()
    if err != nil {
        t.Fatalf("New() returned error: %v", err)
    }
    if app == nil {
        t.Fatal("New() returned nil")
    }
}
```

**Step 2: Update New() signature**

```go
func New(opts ...Option) (*App, error) {
    app := &App{
        ready:           make(chan struct{}),
        shutdownTimeout: 30 * time.Second,
    }
    for _, opt := range opts {
        opt(app)
    }
    return app, nil  // 暂时总是返回 nil error，后续配置解析时可能返回 error
}
```

**Step 3: Run tests**

Run: `go test -v bingo/pkg/app`
Expected: PASS

**Step 4: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): update New() to return error for future config parsing"
```

---

### Task 13: Add Init() method with sync.Once

**Files:**
- Modify: `pkg/app/app.go`
- Modify: `pkg/app/app_test.go`

**Step 1: Write the test**

```go
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

    app.Init()
    app.Init()
    app.Init()

    if initCount != 1 {
        t.Fatalf("Init was called %d times, want 1", initCount)
    }
}
```

**Step 2: Implement Init()**

```go
// Add to App struct:
    initOnce sync.Once
    initErr  error

// Init initializes dependencies (DB, Cache, etc).
// Safe to call multiple times - only executes once.
func (app *App) Init() error {
    app.initOnce.Do(func() {
        // Future: initialize DB, Cache, Logger based on config
        // For now, just mark as initialized
    })
    return app.initErr
}
```

**Step 3: Run tests**

Run: `go test -v bingo/pkg/app -run TestAppInit`
Expected: PASS

**Step 4: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): add Init() method with sync.Once for idempotent initialization"
```

---

### Task 14: Add Close() method

**Files:**
- Modify: `pkg/app/app.go`
- Modify: `pkg/app/app_test.go`

**Step 1: Write the test**

```go
func TestAppClose(t *testing.T) {
    app, _ := New()
    app.Init()

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
```

**Step 2: Implement Close()**

```go
// Close releases resources (DB connections, etc).
// Safe to call multiple times.
func (app *App) Close() error {
    // Future: close DB, Cache connections
    // For now, just a placeholder
    return nil
}
```

**Step 3: Run tests**

Run: `go test -v bingo/pkg/app -run TestAppClose`
Expected: PASS

**Step 4: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): add Close() method for resource cleanup"
```

---

### Task 15: Update Run() to call Init() automatically

**Files:**
- Modify: `pkg/app/app.go`
- Modify: `pkg/app/app_test.go`

**Step 1: Write the test**

```go
func TestRunCallsInit(t *testing.T) {
    var initCalled bool
    app, _ := New(WithInitFunc(func() error {
        initCalled = true
        return nil
    }))

    ctx, cancel := context.WithCancel(context.Background())
    cancel() // cancel immediately

    app.Run(ctx)

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

    app.Init() // explicit init

    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    app.Run(ctx) // should not init again

    if initCount != 1 {
        t.Fatalf("Init was called %d times, want 1", initCount)
    }
}
```

**Step 2: Update Run()**

```go
func (app *App) Run(ctx context.Context) error {
    // Auto-init if not already done
    if err := app.Init(); err != nil {
        return err
    }

    // ... rest of existing Run() implementation
}
```

**Step 3: Run tests**

Run: `go test -v bingo/pkg/app`
Expected: PASS

**Step 4: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): Run() automatically calls Init()"
```

---

### Task 16: Add WithHealthAddr and HealthServer integration

**Files:**
- Modify: `pkg/app/app.go`
- Modify: `pkg/app/options.go`
- Modify: `pkg/app/app_test.go`

**Step 1: Write the test**

```go
func TestAppWithHealthAddr(t *testing.T) {
    app, _ := New(WithHealthAddr(":0"))

    ctx, cancel := context.WithCancel(context.Background())

    errCh := make(chan error, 1)
    go func() {
        errCh <- app.Run(ctx)
    }()

    <-app.Ready()

    // Health server should be running and ready
    addr := app.HealthAddr()
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
```

**Step 2: Implement**

Add to `pkg/app/options.go`:
```go
func WithHealthAddr(addr string) Option {
    return func(app *App) {
        app.healthAddr = addr
    }
}
```

Add to `pkg/app/app.go`:
```go
// Add to App struct:
    healthAddr   string
    healthServer *server.HealthServer

// Add method:
func (app *App) HealthAddr() string {
    if app.healthServer != nil {
        return app.healthServer.Addr()
    }
    return app.healthAddr
}

// Update Run() to start health server:
func (app *App) Run(ctx context.Context) error {
    if err := app.Init(); err != nil {
        return err
    }

    // Start health server if configured
    if app.healthAddr != "" {
        app.healthServer = server.NewHealthServer(app.healthAddr)
        app.runnables = append([]Runnable{app.healthServer}, app.runnables...)
    }

    // ... rest of Run()

    // Update markReady to set health server ready
}

func (app *App) markReady() {
    app.readyOnce.Do(func() {
        if app.healthServer != nil {
            app.healthServer.SetReady(true)
        }
        close(app.ready)
    })
}
```

**Step 3: Run tests**

Run: `go test -v bingo/pkg/app -run TestAppWithHealthAddr`
Expected: PASS

**Step 4: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): add WithHealthAddr for integrated health server"
```

---

### Task 17: Implement shutdown timeout

**Files:**
- Modify: `pkg/app/app.go`
- Modify: `pkg/app/app_test.go`

**Step 1: Write the test**

```go
func TestAppShutdownTimeout(t *testing.T) {
    app, _ := New(WithShutdownTimeout(100 * time.Millisecond))

    // Add a runnable that never stops
    app.Add(runnableFunc(func(ctx context.Context) error {
        <-ctx.Done()
        time.Sleep(5 * time.Second) // simulate slow shutdown
        return nil
    }))

    ctx, cancel := context.WithCancel(context.Background())

    errCh := make(chan error, 1)
    go func() {
        errCh <- app.Run(ctx)
    }()

    <-app.Ready()
    cancel()

    select {
    case err := <-errCh:
        // Should return timeout error, not wait 5 seconds
        if err == nil {
            t.Log("Run returned nil (shutdown completed within timeout)")
        }
    case <-time.After(500 * time.Millisecond):
        t.Fatal("Run did not return within shutdown timeout")
    }
}
```

**Step 2: Update Run() with timeout**

```go
func (app *App) Run(ctx context.Context) error {
    // ... existing init and start logic ...

    // Wait for completion with timeout
    done := make(chan error, 1)
    go func() {
        done <- g.Wait()
    }()

    select {
    case err := <-done:
        return err
    case <-time.After(app.shutdownTimeout):
        return errors.New("shutdown timeout exceeded")
    }
}
```

**Step 3: Run tests**

Run: `go test -v bingo/pkg/app -run TestAppShutdownTimeout`
Expected: PASS

**Step 4: Commit**

```bash
git add pkg/app/
git commit -m "feat(app): implement shutdown timeout"
```

---

## Updated Summary

Phase 2 implementation now includes:

1. **Core Interfaces** (Tasks 1-3): Runnable, Registrar, Named, App basics ✅
2. **Server Adapters** (Tasks 4-6): HTTP, gRPC, WebSocket servers ✅
3. **Health Check** (Task 7): Independent health server ✅
4. **Signals** (Task 8): SetupSignalHandler ✅
5. **Integration** (Tasks 9-10): Dependency access ✅
6. **Migration** (Task 11): Documentation ✅
7. **Complete Lifecycle** (Tasks 12-17): **NEW**
   - Task 12: New() returns error
   - Task 13: Init() with sync.Once
   - Task 14: Close() for cleanup
   - Task 15: Run() calls Init() automatically
   - Task 16: WithHealthAddr integration
   - Task 17: Shutdown timeout

Total: 17 tasks
