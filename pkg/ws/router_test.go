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
	r.Handle("test.method", func(mc *Context) *jsonrpc.Response {
		called = true
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	})

	mc := &Context{
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

	mc := &Context{
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
		return func(mc *Context) *jsonrpc.Response {
			middlewareCalled = true
			return next(mc)
		}
	})

	r.Handle("test", func(mc *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	})

	mc := &Context{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	r.Dispatch(mc)

	assert.True(t, middlewareCalled)
}

func TestRouter_HandlerMiddleware(t *testing.T) {
	r := NewRouter()

	var order []string

	globalMw := func(next Handler) Handler {
		return func(mc *Context) *jsonrpc.Response {
			order = append(order, "global")
			return next(mc)
		}
	}

	handlerMw := func(next Handler) Handler {
		return func(mc *Context) *jsonrpc.Response {
			order = append(order, "handler-mw")
			return next(mc)
		}
	}

	r.Use(globalMw)
	r.Handle("test", func(mc *Context) *jsonrpc.Response {
		order = append(order, "handler")
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}, handlerMw)

	mc := &Context{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	r.Dispatch(mc)

	assert.Equal(t, []string{"global", "handler-mw", "handler"}, order)
}

func TestRouter_Methods(t *testing.T) {
	r := NewRouter()

	r.Handle("method.a", func(mc *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "a")
	})
	r.Handle("method.b", func(mc *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "b")
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
		return func(mc *Context) *jsonrpc.Response {
			globalCalled = true
			return next(mc)
		}
	})

	g := r.Group(func(next Handler) Handler {
		return func(mc *Context) *jsonrpc.Response {
			groupCalled = true
			return next(mc)
		}
	})

	g.Handle("group.method", func(mc *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	})

	mc := &Context{
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
		return func(mc *Context) *jsonrpc.Response {
			authCalled = true
			return next(mc)
		}
	}

	// Public group (no auth)
	public := r.Group()
	public.Handle("public.method", func(mc *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "public")
	})

	// Private group (with auth)
	private := r.Group(authMiddleware)
	private.Handle("private.method", func(mc *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "private")
	})

	// Call public method
	authCalled = false
	mc := &Context{
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

func TestGroup_Use(t *testing.T) {
	r := NewRouter()

	var order []string

	g := r.Group()
	g.Use(func(next Handler) Handler {
		return func(mc *Context) *jsonrpc.Response {
			order = append(order, "group-mw-1")
			return next(mc)
		}
	})
	g.Use(func(next Handler) Handler {
		return func(mc *Context) *jsonrpc.Response {
			order = append(order, "group-mw-2")
			return next(mc)
		}
	})

	g.Handle("test", func(mc *Context) *jsonrpc.Response {
		order = append(order, "handler")
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	})

	mc := &Context{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	r.Dispatch(mc)

	assert.Equal(t, []string{"group-mw-1", "group-mw-2", "handler"}, order)
}
