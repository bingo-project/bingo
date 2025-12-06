// ABOUTME: Tests for WebSocket method router.
// ABOUTME: Verifies routing, middleware application, and method not found handling.

package ws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
)

func TestRouter_Handle(t *testing.T) {
	r := NewRouter()

	called := false
	r.Handle("test.method", func(c *Context) *jsonrpc.Response {
		called = true
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	})

	c := &Context{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test.method"},
		Method:  "test.method",
	}

	resp := r.Dispatch(c)

	assert.True(t, called)
	assert.Nil(t, resp.Error)
	assert.Equal(t, "ok", resp.Result)
}

func TestRouter_MethodNotFound(t *testing.T) {
	r := NewRouter()

	c := &Context{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "unknown"},
		Method:  "unknown",
	}

	resp := r.Dispatch(c)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, "MethodNotFound", resp.Error.Reason)
}

func TestRouter_GlobalMiddleware(t *testing.T) {
	r := NewRouter()

	var middlewareCalled bool
	r.Use(func(next Handler) Handler {
		return func(c *Context) *jsonrpc.Response {
			middlewareCalled = true
			return next(c)
		}
	})

	r.Handle("test", func(c *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	})

	c := &Context{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	r.Dispatch(c)

	assert.True(t, middlewareCalled)
}

func TestRouter_HandlerMiddleware(t *testing.T) {
	r := NewRouter()

	var order []string

	globalMw := func(next Handler) Handler {
		return func(c *Context) *jsonrpc.Response {
			order = append(order, "global")
			return next(c)
		}
	}

	handlerMw := func(next Handler) Handler {
		return func(c *Context) *jsonrpc.Response {
			order = append(order, "handler-mw")
			return next(c)
		}
	}

	r.Use(globalMw)
	r.Handle("test", func(c *Context) *jsonrpc.Response {
		order = append(order, "handler")
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}, handlerMw)

	c := &Context{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	r.Dispatch(c)

	assert.Equal(t, []string{"global", "handler-mw", "handler"}, order)
}

func TestRouter_Methods(t *testing.T) {
	r := NewRouter()

	r.Handle("method.a", func(c *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "a")
	})
	r.Handle("method.b", func(c *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "b")
	})

	methods := r.Methods()

	assert.Len(t, methods, 2)
	assert.Contains(t, methods, "method.a")
	assert.Contains(t, methods, "method.b")
}

func TestRouter_Group(t *testing.T) {
	r := NewRouter()

	var globalCalled, groupCalled bool

	r.Use(func(next Handler) Handler {
		return func(c *Context) *jsonrpc.Response {
			globalCalled = true
			return next(c)
		}
	})

	g := r.Group(func(next Handler) Handler {
		return func(c *Context) *jsonrpc.Response {
			groupCalled = true
			return next(c)
		}
	})

	g.Handle("group.method", func(c *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	})

	c := &Context{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "group.method"},
		Method:  "group.method",
	}

	r.Dispatch(c)

	assert.True(t, globalCalled, "global middleware should be called")
	assert.True(t, groupCalled, "group middleware should be called")
}

func TestRouter_GroupIsolation(t *testing.T) {
	r := NewRouter()

	var authCalled bool
	authMiddleware := func(next Handler) Handler {
		return func(c *Context) *jsonrpc.Response {
			authCalled = true
			return next(c)
		}
	}

	// Public group (no auth)
	public := r.Group()
	public.Handle("public.method", func(c *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "public")
	})

	// Private group (with auth)
	private := r.Group(authMiddleware)
	private.Handle("private.method", func(c *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "private")
	})

	// Call public method
	authCalled = false
	c := &Context{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "public.method"},
		Method:  "public.method",
	}
	r.Dispatch(c)
	assert.False(t, authCalled, "auth should not be called for public method")

	// Call private method
	c.Method = "private.method"
	c.Request.Method = "private.method"
	r.Dispatch(c)
	assert.True(t, authCalled, "auth should be called for private method")
}

func TestGroup_Use(t *testing.T) {
	r := NewRouter()

	var order []string

	g := r.Group()
	g.Use(func(next Handler) Handler {
		return func(c *Context) *jsonrpc.Response {
			order = append(order, "group-mw-1")
			return next(c)
		}
	})
	g.Use(func(next Handler) Handler {
		return func(c *Context) *jsonrpc.Response {
			order = append(order, "group-mw-2")
			return next(c)
		}
	})

	g.Handle("test", func(c *Context) *jsonrpc.Response {
		order = append(order, "handler")
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	})

	c := &Context{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	r.Dispatch(c)

	assert.Equal(t, []string{"group-mw-1", "group-mw-2", "handler"}, order)
}
