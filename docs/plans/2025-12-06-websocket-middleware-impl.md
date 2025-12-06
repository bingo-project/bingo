# WebSocket Middleware Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement middleware system for WebSocket that mirrors HTTP/gRPC patterns with group-based routing.

**Architecture:** Create Router and Group types in `pkg/ws/` that wrap handlers with middleware chains. Client.handleMessage() delegates to Router.Dispatch(). Built-in middlewares in `pkg/ws/middleware/`.

**Tech Stack:** Go, gorilla/websocket, golang.org/x/time/rate (rate limiting)

---

## Task 1: Define Core Types

**Files:**
- Create: `pkg/ws/middleware.go`
- Test: `pkg/ws/middleware_test.go`

**Step 1: Write the failing test**

```go
// pkg/ws/middleware_test.go
package ws

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
)

func TestMiddlewareContext_RequestID(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, requestIDKey{}, "test-123")

	mc := &MiddlewareContext{
		Ctx:       ctx,
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Method:    "test",
		StartTime: time.Now(),
	}

	assert.Equal(t, "test-123", mc.RequestID())
}

func TestMiddlewareChain(t *testing.T) {
	var order []string

	m1 := func(next Handler) Handler {
		return func(mc *MiddlewareContext) *jsonrpc.Response {
			order = append(order, "m1-before")
			resp := next(mc)
			order = append(order, "m1-after")
			return resp
		}
	}

	m2 := func(next Handler) Handler {
		return func(mc *MiddlewareContext) *jsonrpc.Response {
			order = append(order, "m2-before")
			resp := next(mc)
			order = append(order, "m2-after")
			return resp
		}
	}

	handler := func(mc *MiddlewareContext) *jsonrpc.Response {
		order = append(order, "handler")
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	chain := Chain(m1, m2)
	wrapped := chain(handler)

	mc := &MiddlewareContext{
		Request: &jsonrpc.Request{ID: 1},
	}
	wrapped(mc)

	assert.Equal(t, []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}, order)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/... -run TestMiddleware -v`
Expected: FAIL - types not defined

**Step 3: Write minimal implementation**

```go
// pkg/ws/middleware.go
// ABOUTME: Middleware types for WebSocket message handling.
// ABOUTME: Provides middleware chain composition similar to HTTP/gRPC patterns.

package ws

import (
	"context"
	"time"

	"bingo/pkg/jsonrpc"
)

// requestIDKey is the context key for request ID (same as contextx).
type requestIDKey struct{}

// userIDKey is the context key for user ID (same as contextx).
type userIDKey struct{}

// MiddlewareContext contains all information needed by middleware.
type MiddlewareContext struct {
	Ctx       context.Context
	Request   *jsonrpc.Request
	Client    *Client
	Method    string
	StartTime time.Time
}

// RequestID returns the request ID from context.
func (mc *MiddlewareContext) RequestID() string {
	if mc.Ctx == nil {
		return ""
	}
	rid, _ := mc.Ctx.Value(requestIDKey{}).(string)
	return rid
}

// UserID returns the user ID from context.
func (mc *MiddlewareContext) UserID() string {
	if mc.Ctx == nil {
		return ""
	}
	uid, _ := mc.Ctx.Value(userIDKey{}).(string)
	return uid
}

// Handler is a message handler function.
type Handler func(*MiddlewareContext) *jsonrpc.Response

// Middleware wraps a handler with additional functionality.
type Middleware func(Handler) Handler

// Chain combines multiple middlewares into a single middleware.
func Chain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/... -run TestMiddleware -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/middleware.go pkg/ws/middleware_test.go
git commit -m "feat(ws): add middleware types and chain composition"
```

---

## Task 2: Implement Router

**Files:**
- Create: `pkg/ws/router.go`
- Test: `pkg/ws/router_test.go`

**Step 1: Write the failing test**

```go
// pkg/ws/router_test.go
package ws

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
)

func TestRouter_Handle(t *testing.T) {
	r := NewRouter()

	called := false
	r.Handle("test.method", func(mc *MiddlewareContext) *jsonrpc.Response {
		called = true
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	})

	mc := &MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test.method"},
		Method:  "test.method",
	}

	resp := r.Dispatch(mc)

	assert.True(t, called)
	assert.Nil(t, resp.Error)
	assert.Equal(t, "ok", resp.Result)
}

func TestRouter_MethodNotFound(t *testing.T) {
	r := NewRouter()

	mc := &MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "unknown"},
		Method:  "unknown",
	}

	resp := r.Dispatch(mc)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, "MethodNotFound", resp.Error.Reason)
}

func TestRouter_GlobalMiddleware(t *testing.T) {
	r := NewRouter()

	var middlewareCalled bool
	r.Use(func(next Handler) Handler {
		return func(mc *MiddlewareContext) *jsonrpc.Response {
			middlewareCalled = true
			return next(mc)
		}
	})

	r.Handle("test", func(mc *MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	})

	mc := &MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	r.Dispatch(mc)

	assert.True(t, middlewareCalled)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/... -run TestRouter -v`
Expected: FAIL - Router not defined

**Step 3: Write minimal implementation**

```go
// pkg/ws/router.go
// ABOUTME: WebSocket method router with middleware support.
// ABOUTME: Provides group-based routing similar to Gin.

package ws

import (
	"sync"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
)

// Router routes JSON-RPC methods to handlers with middleware.
type Router struct {
	mu          sync.RWMutex
	middlewares []Middleware
	handlers    map[string]*handlerEntry
}

type handlerEntry struct {
	handler     Handler
	middlewares []Middleware
	compiled    Handler // cached compiled handler chain
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]*handlerEntry),
	}
}

// Use adds global middleware that applies to all handlers.
func (r *Router) Use(middlewares ...Middleware) *Router {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, middlewares...)
	// Invalidate compiled handlers
	for _, entry := range r.handlers {
		entry.compiled = nil
	}
	return r
}

// Handle registers a handler for a method.
func (r *Router) Handle(method string, handler Handler, middlewares ...Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[method] = &handlerEntry{
		handler:     handler,
		middlewares: middlewares,
	}
}

// Dispatch routes a request to its handler.
func (r *Router) Dispatch(mc *MiddlewareContext) *jsonrpc.Response {
	r.mu.RLock()
	entry, ok := r.handlers[mc.Method]
	if !ok {
		r.mu.RUnlock()
		return jsonrpc.NewErrorResponse(mc.Request.ID,
			errorsx.New(404, "MethodNotFound", "Method not found: %s", mc.Method))
	}

	// Compile handler chain if not cached
	if entry.compiled == nil {
		r.mu.RUnlock()
		r.mu.Lock()
		if entry.compiled == nil {
			all := append(r.middlewares, entry.middlewares...)
			entry.compiled = Chain(all...)(entry.handler)
		}
		r.mu.Unlock()
		r.mu.RLock()
	}

	compiled := entry.compiled
	r.mu.RUnlock()

	return compiled(mc)
}

// Methods returns all registered method names.
func (r *Router) Methods() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	methods := make([]string, 0, len(r.handlers))
	for method := range r.handlers {
		methods = append(methods, method)
	}
	return methods
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/... -run TestRouter -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/router.go pkg/ws/router_test.go
git commit -m "feat(ws): add Router with method dispatch and middleware"
```

---

## Task 3: Implement Group

**Files:**
- Modify: `pkg/ws/router.go`
- Test: `pkg/ws/router_test.go`

**Step 1: Write the failing test**

```go
// Add to pkg/ws/router_test.go

func TestRouter_Group(t *testing.T) {
	r := NewRouter()

	var globalCalled, groupCalled bool

	r.Use(func(next Handler) Handler {
		return func(mc *MiddlewareContext) *jsonrpc.Response {
			globalCalled = true
			return next(mc)
		}
	})

	g := r.Group(func(next Handler) Handler {
		return func(mc *MiddlewareContext) *jsonrpc.Response {
			groupCalled = true
			return next(mc)
		}
	})

	g.Handle("group.method", func(mc *MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	})

	mc := &MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "group.method"},
		Method:  "group.method",
	}

	r.Dispatch(mc)

	assert.True(t, globalCalled, "global middleware should be called")
	assert.True(t, groupCalled, "group middleware should be called")
}

func TestRouter_GroupIsolation(t *testing.T) {
	r := NewRouter()

	var authCalled bool
	authMiddleware := func(next Handler) Handler {
		return func(mc *MiddlewareContext) *jsonrpc.Response {
			authCalled = true
			return next(mc)
		}
	}

	// Public group (no auth)
	public := r.Group()
	public.Handle("public.method", func(mc *MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "public")
	})

	// Private group (with auth)
	private := r.Group(authMiddleware)
	private.Handle("private.method", func(mc *MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "private")
	})

	// Call public method
	authCalled = false
	mc := &MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "public.method"},
		Method:  "public.method",
	}
	r.Dispatch(mc)
	assert.False(t, authCalled, "auth should not be called for public method")

	// Call private method
	mc.Method = "private.method"
	mc.Request.Method = "private.method"
	r.Dispatch(mc)
	assert.True(t, authCalled, "auth should be called for private method")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/... -run TestRouter_Group -v`
Expected: FAIL - Group not defined

**Step 3: Write minimal implementation**

```go
// Add to pkg/ws/router.go

// Group is a collection of handlers with shared middleware.
type Group struct {
	router      *Router
	middlewares []Middleware
}

// Group creates a new handler group with additional middleware.
func (r *Router) Group(middlewares ...Middleware) *Group {
	return &Group{
		router:      r,
		middlewares: middlewares,
	}
}

// Use adds middleware to this group.
func (g *Group) Use(middlewares ...Middleware) *Group {
	g.middlewares = append(g.middlewares, middlewares...)
	return g
}

// Handle registers a handler in this group.
func (g *Group) Handle(method string, handler Handler, middlewares ...Middleware) {
	// Combine group middlewares with handler-specific middlewares
	all := append(g.middlewares, middlewares...)
	g.router.Handle(method, handler, all...)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/... -run TestRouter_Group -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/router.go pkg/ws/router_test.go
git commit -m "feat(ws): add Group for middleware isolation"
```

---

## Task 4: Implement Recovery Middleware

**Files:**
- Create: `pkg/ws/middleware/recovery.go`
- Test: `pkg/ws/middleware/recovery_test.go`

**Step 1: Write the failing test**

```go
// pkg/ws/middleware/recovery_test.go
package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

func TestRecovery(t *testing.T) {
	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		panic("test panic")
	}

	wrapped := Recovery(handler)

	mc := &ws.MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	resp := wrapped(mc)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, "InternalError", resp.Error.Reason)
	assert.Contains(t, resp.Error.Message, "panic")
}

func TestRecovery_NoError(t *testing.T) {
	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	wrapped := Recovery(handler)

	mc := &ws.MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	resp := wrapped(mc)

	assert.Nil(t, resp.Error)
	assert.Equal(t, "ok", resp.Result)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/middleware/... -run TestRecovery -v`
Expected: FAIL - package/function not found

**Step 3: Write minimal implementation**

```go
// pkg/ws/middleware/recovery.go
// ABOUTME: Panic recovery middleware for WebSocket handlers.
// ABOUTME: Catches panics and returns a JSON-RPC error response.

package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/bingo-project/component-base/log"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// Recovery catches panics and returns an error response.
func Recovery(next ws.Handler) ws.Handler {
	return func(mc *ws.MiddlewareContext) (resp *jsonrpc.Response) {
		defer func() {
			if r := recover(); r != nil {
				log.C(mc.Ctx).Errorw("WebSocket panic recovered",
					"method", mc.Method,
					"panic", r,
					"stack", string(debug.Stack()),
				)
				resp = jsonrpc.NewErrorResponse(mc.Request.ID,
					errorsx.New(500, "InternalError", "panic: %v", r))
			}
		}()
		return next(mc)
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/middleware/... -run TestRecovery -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/middleware/recovery.go pkg/ws/middleware/recovery_test.go
git commit -m "feat(ws): add Recovery middleware"
```

---

## Task 5: Implement RequestID Middleware

**Files:**
- Create: `pkg/ws/middleware/requestid.go`
- Test: `pkg/ws/middleware/requestid_test.go`

**Step 1: Write the failing test**

```go
// pkg/ws/middleware/requestid_test.go
package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

func TestRequestID_UsesClientID(t *testing.T) {
	var capturedRequestID string

	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		capturedRequestID = mc.RequestID()
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	wrapped := RequestID(handler)

	mc := &ws.MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: "client-123", Method: "test"},
		Method:  "test",
	}

	wrapped(mc)

	assert.Equal(t, "client-123", capturedRequestID)
}

func TestRequestID_GeneratesIfMissing(t *testing.T) {
	var capturedRequestID string

	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		capturedRequestID = mc.RequestID()
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	wrapped := RequestID(handler)

	mc := &ws.MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{Method: "test"}, // No ID
		Method:  "test",
	}

	wrapped(mc)

	assert.NotEmpty(t, capturedRequestID)
	assert.Len(t, capturedRequestID, 36) // UUID length
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/middleware/... -run TestRequestID -v`
Expected: FAIL - RequestID not defined

**Step 3: Write minimal implementation**

```go
// pkg/ws/middleware/requestid.go
// ABOUTME: Request ID middleware for WebSocket handlers.
// ABOUTME: Uses client-provided ID or generates UUID.

package middleware

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// requestIDKey is the context key for request ID.
type requestIDKey struct{}

// RequestID adds request ID to context.
// Uses client-provided ID if present, otherwise generates UUID.
func RequestID(next ws.Handler) ws.Handler {
	return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		requestID := ""
		if mc.Request.ID != nil {
			requestID = fmt.Sprintf("%v", mc.Request.ID)
		}
		if requestID == "" {
			requestID = uuid.New().String()
		}

		mc.Ctx = context.WithValue(mc.Ctx, requestIDKey{}, requestID)
		return next(mc)
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/middleware/... -run TestRequestID -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/middleware/requestid.go pkg/ws/middleware/requestid_test.go
git commit -m "feat(ws): add RequestID middleware"
```

---

## Task 6: Implement Auth Middleware

**Files:**
- Create: `pkg/ws/middleware/auth.go`
- Test: `pkg/ws/middleware/auth_test.go`

**Step 1: Write the failing test**

```go
// pkg/ws/middleware/auth_test.go
package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

type mockClient struct {
	authenticated bool
	userID        string
}

func (m *mockClient) IsAuthenticated() bool { return m.authenticated }
func (m *mockClient) GetUserID() string     { return m.userID }

func TestAuth_Authenticated(t *testing.T) {
	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	wrapped := Auth(handler)

	client := &ws.Client{}
	client.UserID = "user-123"
	client.Platform = "web"
	client.LoginTime = 1000

	mc := &ws.MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  client,
		Method:  "test",
	}

	resp := wrapped(mc)

	assert.Nil(t, resp.Error)
	assert.Equal(t, "ok", resp.Result)
}

func TestAuth_Unauthenticated(t *testing.T) {
	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	wrapped := Auth(handler)

	client := &ws.Client{} // Not logged in

	mc := &ws.MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  client,
		Method:  "test",
	}

	resp := wrapped(mc)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, "Unauthorized", resp.Error.Reason)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/middleware/... -run TestAuth -v`
Expected: FAIL - Auth not defined

**Step 3: Write minimal implementation**

```go
// pkg/ws/middleware/auth.go
// ABOUTME: Authentication middleware for WebSocket handlers.
// ABOUTME: Blocks unauthenticated requests with 401 error.

package middleware

import (
	"context"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// userIDKey is the context key for user ID.
type userIDKey struct{}

// Auth requires the client to be authenticated.
func Auth(next ws.Handler) ws.Handler {
	return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		if mc.Client == nil || !mc.Client.IsAuthenticated() {
			return jsonrpc.NewErrorResponse(mc.Request.ID,
				errorsx.New(401, "Unauthorized", "Login required"))
		}

		// Add user ID to context
		mc.Ctx = context.WithValue(mc.Ctx, userIDKey{}, mc.Client.UserID)

		return next(mc)
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/middleware/... -run TestAuth -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/middleware/auth.go pkg/ws/middleware/auth_test.go
git commit -m "feat(ws): add Auth middleware"
```

---

## Task 7: Implement Logger Middleware

**Files:**
- Create: `pkg/ws/middleware/logger.go`
- Test: `pkg/ws/middleware/logger_test.go`

**Step 1: Write the failing test**

```go
// pkg/ws/middleware/logger_test.go
package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

func TestLogger_Success(t *testing.T) {
	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	wrapped := Logger(handler)

	mc := &ws.MiddlewareContext{
		Ctx:       context.Background(),
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Method:    "test",
		StartTime: time.Now(),
	}

	resp := wrapped(mc)

	assert.Nil(t, resp.Error)
	// Logger should not modify the response
	assert.Equal(t, "ok", resp.Result)
}

func TestLogger_Error(t *testing.T) {
	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewErrorResponse(mc.Request.ID,
			errorsx.New(400, "BadRequest", "test error"))
	}

	wrapped := Logger(handler)

	mc := &ws.MiddlewareContext{
		Ctx:       context.Background(),
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Method:    "test",
		StartTime: time.Now(),
	}

	resp := wrapped(mc)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, "BadRequest", resp.Error.Reason)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/middleware/... -run TestLogger -v`
Expected: FAIL - Logger not defined

**Step 3: Write minimal implementation**

```go
// pkg/ws/middleware/logger.go
// ABOUTME: Request logging middleware for WebSocket handlers.
// ABOUTME: Logs method, latency, and error status.

package middleware

import (
	"time"

	"github.com/bingo-project/component-base/log"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// Logger logs request details after handling.
func Logger(next ws.Handler) ws.Handler {
	return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		resp := next(mc)

		fields := []any{
			"method", mc.Method,
			"latency", time.Since(mc.StartTime),
		}

		if mc.Client != nil {
			fields = append(fields, "client_addr", mc.Client.Addr)
			if mc.Client.UserID != "" {
				fields = append(fields, "user_id", mc.Client.UserID)
			}
		}

		if resp.Error != nil {
			fields = append(fields, "error", resp.Error.Reason)
			log.C(mc.Ctx).Warnw("WebSocket request failed", fields...)
		} else {
			log.C(mc.Ctx).Infow("WebSocket request", fields...)
		}

		return resp
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/middleware/... -run TestLogger -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/middleware/logger.go pkg/ws/middleware/logger_test.go
git commit -m "feat(ws): add Logger middleware"
```

---

## Task 8: Implement RateLimit Middleware

**Files:**
- Create: `pkg/ws/middleware/ratelimit.go`
- Test: `pkg/ws/middleware/ratelimit_test.go`

**Step 1: Write the failing test**

```go
// pkg/ws/middleware/ratelimit_test.go
package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

func TestRateLimit_Allows(t *testing.T) {
	cfg := &RateLimitConfig{
		Default: 10, // 10 requests per second
	}

	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	wrapped := RateLimit(cfg)(handler)

	client := &ws.Client{}
	mc := &ws.MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  client,
		Method:  "test",
	}

	resp := wrapped(mc)

	assert.Nil(t, resp.Error)
	assert.Equal(t, "ok", resp.Result)
}

func TestRateLimit_Blocks(t *testing.T) {
	cfg := &RateLimitConfig{
		Default: 1, // 1 request per second
	}

	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	wrapped := RateLimit(cfg)(handler)

	client := &ws.Client{}
	mc := &ws.MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  client,
		Method:  "test",
	}

	// First request should succeed
	resp := wrapped(mc)
	assert.Nil(t, resp.Error)

	// Second request immediately should fail
	resp = wrapped(mc)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "TooManyRequests", resp.Error.Reason)
}

func TestRateLimit_MethodSpecific(t *testing.T) {
	cfg := &RateLimitConfig{
		Default: 1,
		Methods: map[string]float64{
			"heartbeat": 0, // No limit
		},
	}

	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	wrapped := RateLimit(cfg)(handler)

	client := &ws.Client{}

	// Heartbeat should always succeed
	mc := &ws.MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "heartbeat"},
		Client:  client,
		Method:  "heartbeat",
	}

	for i := 0; i < 10; i++ {
		resp := wrapped(mc)
		assert.Nil(t, resp.Error, "heartbeat %d should succeed", i)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/middleware/... -run TestRateLimit -v`
Expected: FAIL - RateLimit not defined

**Step 3: Write minimal implementation**

```go
// pkg/ws/middleware/ratelimit.go
// ABOUTME: Rate limiting middleware for WebSocket handlers.
// ABOUTME: Uses token bucket algorithm with per-method configuration.

package middleware

import (
	"sync"

	"golang.org/x/time/rate"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// RateLimitConfig configures rate limiting.
type RateLimitConfig struct {
	Default float64            // Default requests per second (0 = unlimited)
	Methods map[string]float64 // Per-method limits (0 = unlimited)
}

// clientLimiters stores per-client limiters.
type clientLimiters struct {
	mu       sync.RWMutex
	limiters map[*ws.Client]map[string]*rate.Limiter
}

var limiters = &clientLimiters{
	limiters: make(map[*ws.Client]map[string]*rate.Limiter),
}

func (cl *clientLimiters) get(client *ws.Client, method string, limit float64) *rate.Limiter {
	cl.mu.RLock()
	if methods, ok := cl.limiters[client]; ok {
		if limiter, ok := methods[method]; ok {
			cl.mu.RUnlock()
			return limiter
		}
	}
	cl.mu.RUnlock()

	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.limiters[client] == nil {
		cl.limiters[client] = make(map[string]*rate.Limiter)
	}

	if _, ok := cl.limiters[client][method]; !ok {
		cl.limiters[client][method] = rate.NewLimiter(rate.Limit(limit), int(limit)+1)
	}

	return cl.limiters[client][method]
}

func (cl *clientLimiters) remove(client *ws.Client) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	delete(cl.limiters, client)
}

// RateLimit limits request rate per client per method.
func RateLimit(cfg *RateLimitConfig) ws.Middleware {
	return func(next ws.Handler) ws.Handler {
		return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
			if mc.Client == nil {
				return next(mc)
			}

			// Get limit for this method
			limit := cfg.Default
			if methodLimit, ok := cfg.Methods[mc.Method]; ok {
				limit = methodLimit
			}

			// No limit
			if limit == 0 {
				return next(mc)
			}

			// Check rate limit
			limiter := limiters.get(mc.Client, mc.Method, limit)
			if !limiter.Allow() {
				return jsonrpc.NewErrorResponse(mc.Request.ID,
					errorsx.New(429, "TooManyRequests", "Rate limit exceeded"))
			}

			return next(mc)
		}
	}
}

// CleanupClientLimiters removes limiters for a disconnected client.
// Call this when client disconnects.
func CleanupClientLimiters(client *ws.Client) {
	limiters.remove(client)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/middleware/... -run TestRateLimit -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/middleware/ratelimit.go pkg/ws/middleware/ratelimit_test.go
git commit -m "feat(ws): add RateLimit middleware with token bucket"
```

---

## Task 9: Add Client ID Field

**Files:**
- Modify: `pkg/ws/client.go`
- Test: `pkg/ws/client_test.go`

**Step 1: Write the failing test**

```go
// Add to pkg/ws/client_test.go

func TestClient_HasID(t *testing.T) {
	hub := NewHub()
	conn := &websocket.Conn{} // mock
	ctx := context.Background()
	adapter := jsonrpc.NewAdapter()

	client := NewClient(hub, conn, ctx, adapter)

	assert.NotEmpty(t, client.ID)
	assert.Len(t, client.ID, 36) // UUID length
}

func TestClient_IDIsUnique(t *testing.T) {
	hub := NewHub()
	ctx := context.Background()
	adapter := jsonrpc.NewAdapter()

	client1 := NewClient(hub, nil, ctx, adapter)
	client2 := NewClient(hub, nil, ctx, adapter)

	assert.NotEqual(t, client1.ID, client2.ID)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/... -run TestClient_HasID -v`
Expected: FAIL - ID field not found

**Step 3: Write minimal implementation**

```go
// Modify pkg/ws/client.go

// Add import
import "github.com/google/uuid"

// Add ID field to Client struct (after Addr field):
	ID             string // Unique client identifier

// In NewClient function, add:
	c := &Client{
		ID:            uuid.New().String(), // Add this line
		hub:           hub,
		// ... rest unchanged
	}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/... -run TestClient_HasID -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/client.go pkg/ws/client_test.go
git commit -m "feat(ws): add unique ID to Client"
```

---

## Task 10: Extend Hub with Management APIs

**Files:**
- Modify: `pkg/ws/hub.go`
- Test: `pkg/ws/hub_test.go`

**Step 1: Write the failing test**

```go
// Add to pkg/ws/hub_test.go

func TestHub_GetClient(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// Create and register client
	client := &Client{
		ID:   "test-client-id",
		Send: make(chan []byte, 256),
	}
	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	// Get by ID
	found := hub.GetClient("test-client-id")
	assert.Equal(t, client, found)

	// Not found
	notFound := hub.GetClient("unknown")
	assert.Nil(t, notFound)
}

func TestHub_GetClientsByUser(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// Create clients for same user
	client1 := &Client{
		ID:       "c1",
		UserID:   "user-123",
		Platform: "web",
		Send:     make(chan []byte, 256),
	}
	client2 := &Client{
		ID:       "c2",
		UserID:   "user-123",
		Platform: "mobile",
		Send:     make(chan []byte, 256),
	}

	hub.clients[client1] = true
	hub.clients[client2] = true
	hub.users[userKey("web", "user-123")] = client1
	hub.users[userKey("mobile", "user-123")] = client2

	clients := hub.GetClientsByUser("user-123")
	assert.Len(t, clients, 2)
}

func TestHub_Stats(t *testing.T) {
	hub := NewHub()

	// Add some test data
	hub.anonymous[&Client{}] = true
	hub.clients[&Client{Platform: "web"}] = true
	hub.clients[&Client{Platform: "mobile"}] = true

	stats := hub.Stats()

	assert.Equal(t, int64(1), stats.AnonymousConns)
	assert.Equal(t, int64(2), stats.AuthenticatedConns)
	assert.Equal(t, int64(3), stats.TotalConnections)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/... -run "TestHub_GetClient|TestHub_Stats" -v`
Expected: FAIL - methods not defined

**Step 3: Write minimal implementation**

```go
// Add to pkg/ws/hub.go

// Add clientsByID map to Hub struct:
	clientsByID     map[string]*Client
	clientsByIDLock sync.RWMutex

// Update NewHubWithConfig to initialize:
	clientsByID: make(map[string]*Client),

// Update handleRegister to track by ID:
func (h *Hub) handleRegister(client *Client) {
	h.anonymousLock.Lock()
	h.anonymous[client] = true
	h.anonymousLock.Unlock()

	h.clientsByIDLock.Lock()
	h.clientsByID[client.ID] = client
	h.clientsByIDLock.Unlock()

	h.logger.Debugw("WebSocket client connected", "addr", client.Addr, "id", client.ID)
}

// Update handleUnregister to remove from clientsByID:
// Add after removing from users:
	h.clientsByIDLock.Lock()
	delete(h.clientsByID, client.ID)
	h.clientsByIDLock.Unlock()

// Add new methods:

// GetClient returns a client by ID.
func (h *Hub) GetClient(clientID string) *Client {
	h.clientsByIDLock.RLock()
	defer h.clientsByIDLock.RUnlock()
	return h.clientsByID[clientID]
}

// GetClientsByUser returns all clients for a user.
func (h *Hub) GetClientsByUser(userID string) []*Client {
	h.userLock.RLock()
	defer h.userLock.RUnlock()

	var clients []*Client
	suffix := "_" + userID
	for key, client := range h.users {
		if len(key) > len(suffix) && key[len(key)-len(suffix):] == suffix {
			clients = append(clients, client)
		}
	}
	return clients
}

// KickClient disconnects a client by ID.
func (h *Hub) KickClient(clientID string, reason string) bool {
	client := h.GetClient(clientID)
	if client == nil {
		return false
	}
	h.kickClient(client, reason)
	return true
}

// KickUser disconnects all clients for a user.
func (h *Hub) KickUser(userID string, reason string) int {
	clients := h.GetClientsByUser(userID)
	for _, client := range clients {
		h.kickClient(client, reason)
	}
	return len(clients)
}

// HubStats contains hub statistics.
type HubStats struct {
	TotalConnections      int64
	AuthenticatedConns    int64
	AnonymousConns        int64
	ConnectionsByPlatform map[string]int
}

// Stats returns current hub statistics.
func (h *Hub) Stats() *HubStats {
	h.anonymousLock.RLock()
	anonymous := int64(len(h.anonymous))
	h.anonymousLock.RUnlock()

	h.clientsLock.RLock()
	authenticated := int64(len(h.clients))
	byPlatform := make(map[string]int)
	for client := range h.clients {
		if client.Platform != "" {
			byPlatform[client.Platform]++
		}
	}
	h.clientsLock.RUnlock()

	return &HubStats{
		TotalConnections:      anonymous + authenticated,
		AuthenticatedConns:    authenticated,
		AnonymousConns:        anonymous,
		ConnectionsByPlatform: byPlatform,
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/... -run "TestHub_GetClient|TestHub_Stats" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/hub.go pkg/ws/hub_test.go
git commit -m "feat(ws): add Hub management APIs (GetClient, Stats, Kick)"
```

---

## Task 11: Integrate Router into Client

**Files:**
- Modify: `pkg/ws/client.go`
- Modify: `pkg/ws/hub.go`
- Test: Integration test

**Step 1: Write the failing test**

```go
// Add to pkg/ws/client_test.go

func TestClient_UsesRouter(t *testing.T) {
	hub := NewHub()
	router := NewRouter()

	var handlerCalled bool
	router.Handle("test.method", func(mc *MiddlewareContext) *jsonrpc.Response {
		handlerCalled = true
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	})

	client := NewClient(hub, nil, context.Background(), nil,
		WithRouter(router),
	)

	// Simulate message handling
	msg := `{"jsonrpc":"2.0","method":"test.method","id":1}`
	client.handleMessage([]byte(msg))

	assert.True(t, handlerCalled)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/... -run TestClient_UsesRouter -v`
Expected: FAIL - WithRouter not defined

**Step 3: Write minimal implementation**

```go
// Modify pkg/ws/client.go

// Add router field to Client struct:
	router         *Router

// Add option function:
// WithRouter sets a router for the client.
func WithRouter(r *Router) ClientOption {
	return func(c *Client) {
		c.router = r
	}
}

// Modify handleMessage to use router when available:
func (c *Client) handleMessage(data []byte) {
	// ... keep panic recovery ...

	var req jsonrpc.Request
	if err := json.Unmarshal(data, &req); err != nil {
		resp := jsonrpc.NewErrorResponse(nil,
			errorsx.New(400, "ParseError", "Invalid JSON: %s", err.Error()))
		c.sendJSON(resp)
		return
	}

	// Update heartbeat for any message
	c.Heartbeat(time.Now().Unix())

	// Use router if available
	if c.router != nil {
		mc := &MiddlewareContext{
			Ctx:       c.ctx,
			Request:   &req,
			Client:    c,
			Method:    req.Method,
			StartTime: time.Now(),
		}
		resp := c.router.Dispatch(mc)
		c.sendJSON(resp)
		return
	}

	// ... keep legacy handling for backward compatibility ...
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/... -run TestClient_UsesRouter -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/client.go pkg/ws/client_test.go
git commit -m "feat(ws): integrate Router into Client message handling"
```

---

## Task 12: Create Built-in Handlers

**Files:**
- Create: `pkg/ws/handlers.go`
- Test: `pkg/ws/handlers_test.go`

**Step 1: Write the failing test**

```go
// pkg/ws/handlers_test.go
package ws

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
)

func TestHeartbeatHandler(t *testing.T) {
	mc := &MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "heartbeat"},
		Client:  &Client{HeartbeatTime: 0},
		Method:  "heartbeat",
	}

	before := time.Now().Unix()
	resp := HeartbeatHandler(mc)
	after := time.Now().Unix()

	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]any)
	assert.Equal(t, "ok", result["status"])

	serverTime := int64(result["server_time"].(float64))
	assert.GreaterOrEqual(t, serverTime, before)
	assert.LessOrEqual(t, serverTime, after)
}

func TestSubscribeHandler(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	client := &Client{
		ID:        "test",
		UserID:    "user-1",
		Platform:  "web",
		LoginTime: 1000,
		hub:       hub,
		Send:      make(chan []byte, 256),
	}

	params, _ := json.Marshal(map[string][]string{"topics": {"market.BTC"}})
	mc := &MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "subscribe", Params: params},
		Client:  client,
		Method:  "subscribe",
	}

	resp := SubscribeHandler(mc)

	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]any)
	assert.Contains(t, result["subscribed"], "market.BTC")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ws/... -run "TestHeartbeatHandler|TestSubscribeHandler" -v`
Expected: FAIL - handlers not defined

**Step 3: Write minimal implementation**

```go
// pkg/ws/handlers.go
// ABOUTME: Built-in handlers for common WebSocket methods.
// ABOUTME: Provides heartbeat, subscribe, and unsubscribe functionality.

package ws

import (
	"encoding/json"
	"time"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
)

// HeartbeatHandler responds to heartbeat requests.
func HeartbeatHandler(mc *MiddlewareContext) *jsonrpc.Response {
	return jsonrpc.NewResponse(mc.Request.ID, map[string]any{
		"status":      "ok",
		"server_time": time.Now().Unix(),
	})
}

// SubscribeHandler handles topic subscription.
func SubscribeHandler(mc *MiddlewareContext) *jsonrpc.Response {
	var params struct {
		Topics []string `json:"topics"`
	}

	if err := json.Unmarshal(mc.Request.Params, &params); err != nil {
		return jsonrpc.NewErrorResponse(mc.Request.ID,
			errorsx.New(400, "InvalidParams", "Invalid subscribe params"))
	}

	if len(params.Topics) == 0 {
		return jsonrpc.NewErrorResponse(mc.Request.ID,
			errorsx.New(400, "InvalidParams", "Topics required"))
	}

	result := make(chan []string, 1)
	mc.Client.hub.Subscribe <- &SubscribeEvent{
		Client: mc.Client,
		Topics: params.Topics,
		Result: result,
	}

	subscribed := <-result
	return jsonrpc.NewResponse(mc.Request.ID, map[string]any{
		"subscribed": subscribed,
	})
}

// UnsubscribeHandler handles topic unsubscription.
func UnsubscribeHandler(mc *MiddlewareContext) *jsonrpc.Response {
	var params struct {
		Topics []string `json:"topics"`
	}

	if err := json.Unmarshal(mc.Request.Params, &params); err != nil {
		return jsonrpc.NewErrorResponse(mc.Request.ID,
			errorsx.New(400, "InvalidParams", "Invalid unsubscribe params"))
	}

	if len(params.Topics) == 0 {
		return jsonrpc.NewErrorResponse(mc.Request.ID,
			errorsx.New(400, "InvalidParams", "Topics required"))
	}

	mc.Client.hub.Unsubscribe <- &UnsubscribeEvent{
		Client: mc.Client,
		Topics: params.Topics,
	}

	return jsonrpc.NewResponse(mc.Request.ID, map[string]any{
		"unsubscribed": params.Topics,
	})
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ws/... -run "TestHeartbeatHandler|TestSubscribeHandler" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ws/handlers.go pkg/ws/handlers_test.go
git commit -m "feat(ws): add built-in handlers for heartbeat and subscriptions"
```

---

## Task 13: Update apiserver Router Registration

**Files:**
- Modify: `internal/apiserver/router/ws.go`
- Modify: `internal/apiserver/run.go`
- Modify: `internal/apiserver/handler/ws/handler.go`

**Step 1: Update ws.go to use new Router**

```go
// internal/apiserver/router/ws.go
package router

import (
	"bingo/internal/apiserver/biz"
	wshandler "bingo/internal/apiserver/handler/ws"
	"bingo/pkg/api/apiserver/v1"
	"bingo/pkg/ws"
	"bingo/pkg/ws/middleware"
)

// RegisterWSHandlers registers all WebSocket handlers with the router.
func RegisterWSHandlers(router *ws.Router, b biz.IBiz) {
	systemHandler := wshandler.NewSystemHandler(b)
	authHandler := wshandler.NewAuthHandler(b)

	// Global middleware
	router.Use(
		middleware.Recovery,
		middleware.RequestID,
		middleware.RateLimit(&middleware.RateLimitConfig{
			Default: 10,
			Methods: map[string]float64{
				"heartbeat": 0, // No limit
			},
		}),
	)

	// Public methods (no auth)
	public := router.Group()
	public.Handle("heartbeat", ws.HeartbeatHandler)
	public.Handle("system.healthz", wrapBizHandler(systemHandler.Healthz, &struct{}{}))
	public.Handle("system.version", wrapBizHandler(systemHandler.Version, &struct{}{}))
	public.Handle("auth.login", wrapBizHandler(authHandler.Login, &v1.LoginRequest{}))

	// Private methods (require auth)
	private := router.Group(middleware.Auth, middleware.Logger)
	private.Handle("subscribe", ws.SubscribeHandler)
	private.Handle("unsubscribe", ws.UnsubscribeHandler)
	private.Handle("auth.user-info", wrapBizHandler(authHandler.UserInfo, &struct{}{}))
}

// wrapBizHandler adapts a biz handler to ws.Handler.
func wrapBizHandler[T any](handler func(context.Context, *T) (any, error), reqType *T) ws.Handler {
	return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		req := new(T)
		if len(mc.Request.Params) > 0 {
			if err := json.Unmarshal(mc.Request.Params, req); err != nil {
				return jsonrpc.NewErrorResponse(mc.Request.ID,
					errorsx.New(400, "InvalidParams", "Invalid params: %s", err.Error()))
			}
		}

		resp, err := handler(mc.Ctx, req)
		if err != nil {
			return jsonrpc.NewErrorResponse(mc.Request.ID, err)
		}

		return jsonrpc.NewResponse(mc.Request.ID, resp)
	}
}
```

**Step 2: Update run.go**

```go
// Modify initWebSocket in internal/apiserver/run.go

func initWebSocket() (*gin.Engine, *ws.Hub) {
	hub := ws.NewHub()

	// Create router and register handlers
	router := ws.NewRouter()
	bizInstance := biz.NewBiz(store.S)
	router.RegisterWSHandlers(router, bizInstance)

	// Create Gin engine for WebSocket
	engine := bootstrap.InitGinForWebSocket()

	// Register WebSocket route
	handler := wshandler.NewHandler(hub, router, facade.Config.WebSocket)
	engine.GET("/ws", handler.ServeWS)

	return engine, hub
}
```

**Step 3: Update handler.go**

```go
// Modify internal/apiserver/handler/ws/handler.go

type Handler struct {
	hub      *ws.Hub
	router   *ws.Router
	upgrader websocket.Upgrader
}

func NewHandler(hub *ws.Hub, router *ws.Router, cfg *config.WebSocket) *Handler {
	// ... update to use router instead of adapter
}

func (h *Handler) ServeWS(c *gin.Context) {
	// ... update to pass router to NewClient
	client := ws.NewClient(h.hub, conn, ctx,
		ws.WithRouter(h.router),
		ws.WithTokenParser(tokenParser),
		ws.WithContextUpdater(contextUpdater),
	)
	// ...
}
```

**Step 4: Run tests**

Run: `go test ./internal/apiserver/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/apiserver/router/ws.go internal/apiserver/run.go internal/apiserver/handler/ws/handler.go
git commit -m "refactor(ws): migrate to Router-based handler registration"
```

---

## Task 14: Add Integration Tests

**Files:**
- Create: `pkg/ws/integration_test.go`

**Step 1: Write integration test**

```go
// pkg/ws/integration_test.go
package ws_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
	"bingo/pkg/ws/middleware"
)

func TestFullMiddlewareChain(t *testing.T) {
	// Setup
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	router := ws.NewRouter()

	// Add middlewares
	var order []string
	router.Use(
		func(next ws.Handler) ws.Handler {
			return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
				order = append(order, "recovery")
				return middleware.Recovery(next)(mc)
			}
		},
		func(next ws.Handler) ws.Handler {
			return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
				order = append(order, "requestid")
				return middleware.RequestID(next)(mc)
			}
		},
	)

	// Public handler
	router.Handle("public.test", func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		order = append(order, "handler")
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	})

	// Private handler
	private := router.Group(middleware.Auth)
	private.Handle("private.test", func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		order = append(order, "private-handler")
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	})

	// Test public method
	client := &ws.Client{
		ID:   "test",
		Send: make(chan []byte, 256),
	}

	mc := &ws.MiddlewareContext{
		Ctx:       context.Background(),
		Request:   &jsonrpc.Request{ID: 1, Method: "public.test"},
		Client:    client,
		Method:    "public.test",
		StartTime: time.Now(),
	}

	resp := router.Dispatch(mc)
	assert.Nil(t, resp.Error)
	assert.Equal(t, []string{"recovery", "requestid", "handler"}, order)

	// Test private method without auth
	order = nil
	mc.Method = "private.test"
	mc.Request.Method = "private.test"
	resp = router.Dispatch(mc)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "Unauthorized", resp.Error.Reason)

	// Test private method with auth
	order = nil
	client.UserID = "user-1"
	client.Platform = "web"
	client.LoginTime = 1000
	resp = router.Dispatch(mc)
	assert.Nil(t, resp.Error)
	assert.Contains(t, order, "private-handler")
}
```

**Step 2: Run integration tests**

Run: `go test ./pkg/ws/... -run TestFullMiddlewareChain -v`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/ws/integration_test.go
git commit -m "test(ws): add middleware chain integration test"
```

---

## Task 15: Update Documentation

**Files:**
- Modify: `docs/plans/2025-12-06-websocket-middleware-design.md`

**Step 1: Add implementation notes**

Add section at the end of design doc:

```markdown
## Implementation Notes

Completed implementation:
- [x] Core types (MiddlewareContext, Handler, Middleware)
- [x] Router with group support
- [x] Built-in middlewares (Recovery, RequestID, Auth, Logger, RateLimit)
- [x] Hub management APIs (GetClient, Stats, Kick)
- [x] Client router integration
- [x] Built-in handlers (heartbeat, subscribe, unsubscribe)
- [x] apiserver migration

Key files:
- `pkg/ws/middleware.go` - Core types
- `pkg/ws/router.go` - Router and Group
- `pkg/ws/middleware/*.go` - Built-in middlewares
- `pkg/ws/handlers.go` - Built-in handlers
```

**Step 2: Commit**

```bash
git add docs/plans/2025-12-06-websocket-middleware-design.md
git commit -m "docs(ws): update design doc with implementation status"
```

---

## Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Core types | `pkg/ws/middleware.go` |
| 2 | Router | `pkg/ws/router.go` |
| 3 | Group | `pkg/ws/router.go` |
| 4 | Recovery middleware | `pkg/ws/middleware/recovery.go` |
| 5 | RequestID middleware | `pkg/ws/middleware/requestid.go` |
| 6 | Auth middleware | `pkg/ws/middleware/auth.go` |
| 7 | Logger middleware | `pkg/ws/middleware/logger.go` |
| 8 | RateLimit middleware | `pkg/ws/middleware/ratelimit.go` |
| 9 | Client ID field | `pkg/ws/client.go` |
| 10 | Hub management APIs | `pkg/ws/hub.go` |
| 11 | Router integration | `pkg/ws/client.go` |
| 12 | Built-in handlers | `pkg/ws/handlers.go` |
| 13 | apiserver migration | `internal/apiserver/` |
| 14 | Integration tests | `pkg/ws/integration_test.go` |
| 15 | Documentation | `docs/plans/` |
