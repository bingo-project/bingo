// ABOUTME: Integration tests for WebSocket middleware chain.
// ABOUTME: Validates end-to-end middleware execution order and group isolation.

package ws_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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

	// Add middlewares with tracking
	var order []string
	router.Use(
		func(next ws.Handler) ws.Handler {
			return func(c *ws.Context) *jsonrpc.Response {
				order = append(order, "recovery")

				return middleware.Recovery(next)(c)
			}
		},
		func(next ws.Handler) ws.Handler {
			return func(c *ws.Context) *jsonrpc.Response {
				order = append(order, "requestid")

				return middleware.RequestID(next)(c)
			}
		},
	)

	// Public handler
	router.Handle("public.test", func(c *ws.Context) *jsonrpc.Response {
		order = append(order, "handler")

		return jsonrpc.NewResponse(c.Request.ID, "ok")
	})

	// Private handler with auth middleware
	private := router.Group(middleware.Auth)
	private.Handle("private.test", func(c *ws.Context) *jsonrpc.Response {
		order = append(order, "private-handler")

		return jsonrpc.NewResponse(c.Request.ID, "ok")
	})

	// Test public method - should pass through all middlewares
	client := &ws.Client{
		ID:   "test",
		Send: make(chan []byte, 256),
	}

	c := &ws.Context{
		Context:   context.Background(),
		Request:   &jsonrpc.Request{ID: 1, Method: "public.test"},
		Client:    client,
		Method:    "public.test",
		StartTime: time.Now(),
	}

	resp := router.Dispatch(c)
	assert.Nil(t, resp.Error)
	assert.Equal(t, []string{"recovery", "requestid", "handler"}, order)

	// Test private method without auth - should be rejected
	order = nil
	c.Method = "private.test"
	c.Request.Method = "private.test"
	resp = router.Dispatch(c)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "Unauthorized", resp.Error.Reason)

	// Test private method with auth - should pass
	order = nil
	client.UserID = "user-1"
	client.Platform = "web"
	client.LoginTime = 1000
	resp = router.Dispatch(c)
	assert.Nil(t, resp.Error)
	assert.Contains(t, order, "private-handler")
}

func TestMiddlewareChain_ExecutionOrder(t *testing.T) {
	router := ws.NewRouter()

	var order []string

	// Add numbered middlewares to verify order
	for i := 1; i <= 3; i++ {
		n := i // capture
		router.Use(func(next ws.Handler) ws.Handler {
			return func(c *ws.Context) *jsonrpc.Response {
				order = append(order, "before-"+string(rune('0'+n)))
				resp := next(c)
				order = append(order, "after-"+string(rune('0'+n)))

				return resp
			}
		})
	}

	router.Handle("test", func(c *ws.Context) *jsonrpc.Response {
		order = append(order, "handler")

		return jsonrpc.NewResponse(c.Request.ID, "ok")
	})

	c := &ws.Context{
		Context:   context.Background(),
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Method:    "test",
		StartTime: time.Now(),
	}

	router.Dispatch(c)

	// Verify onion execution: before1 -> before2 -> before3 -> handler -> after3 -> after2 -> after1
	expected := []string{
		"before-1", "before-2", "before-3",
		"handler",
		"after-3", "after-2", "after-1",
	}
	assert.Equal(t, expected, order)
}

func TestRateLimitIntegration(t *testing.T) {
	router := ws.NewRouter()

	// Create rate limiter store
	store := middleware.NewRateLimiterStore()

	// Apply rate limit of 1 request per second
	router.Use(middleware.RateLimitWithStore(&middleware.RateLimitConfig{
		Default: 1,
	}, store))

	router.Handle("test", func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	})

	client := &ws.Client{
		ID:   "test-client",
		Send: make(chan []byte, 256),
	}

	c := &ws.Context{
		Context:   context.Background(),
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Client:    client,
		Method:    "test",
		StartTime: time.Now(),
	}

	// First request should pass
	resp := router.Dispatch(c)
	assert.Nil(t, resp.Error)

	// Second request should pass (burst allows 2)
	resp = router.Dispatch(c)
	assert.Nil(t, resp.Error)

	// Third request should be rate limited
	resp = router.Dispatch(c)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "TooManyRequests", resp.Error.Reason)

	// Clean up
	store.Remove(client)
}

func TestGroupMiddlewareIsolation(t *testing.T) {
	router := ws.NewRouter()

	var publicMiddlewareCalled, privateMiddlewareCalled bool

	// Global middleware
	router.Use(func(next ws.Handler) ws.Handler {
		return func(c *ws.Context) *jsonrpc.Response {
			publicMiddlewareCalled = true

			return next(c)
		}
	})

	// Public group (no additional middleware)
	public := router.Group()
	public.Handle("public.test", func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "public")
	})

	// Private group with additional middleware
	private := router.Group(func(next ws.Handler) ws.Handler {
		return func(c *ws.Context) *jsonrpc.Response {
			privateMiddlewareCalled = true

			return next(c)
		}
	})
	private.Handle("private.test", func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "private")
	})

	// Test public method - only global middleware should be called
	c := &ws.Context{
		Context:   context.Background(),
		Request:   &jsonrpc.Request{ID: 1, Method: "public.test"},
		Method:    "public.test",
		StartTime: time.Now(),
	}

	publicMiddlewareCalled = false
	privateMiddlewareCalled = false
	router.Dispatch(c)
	assert.True(t, publicMiddlewareCalled, "global middleware should be called")
	assert.False(t, privateMiddlewareCalled, "private middleware should NOT be called for public method")

	// Test private method - both middlewares should be called
	c.Method = "private.test"
	c.Request.Method = "private.test"
	publicMiddlewareCalled = false
	privateMiddlewareCalled = false
	router.Dispatch(c)
	assert.True(t, publicMiddlewareCalled, "global middleware should be called")
	assert.True(t, privateMiddlewareCalled, "private middleware should be called for private method")
}
