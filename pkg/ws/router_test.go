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

func TestRouter_HandlerMiddleware(t *testing.T) {
	r := NewRouter()

	var order []string

	globalMw := func(next Handler) Handler {
		return func(mc *MiddlewareContext) *jsonrpc.Response {
			order = append(order, "global")
			return next(mc)
		}
	}

	handlerMw := func(next Handler) Handler {
		return func(mc *MiddlewareContext) *jsonrpc.Response {
			order = append(order, "handler-mw")
			return next(mc)
		}
	}

	r.Use(globalMw)
	r.Handle("test", func(mc *MiddlewareContext) *jsonrpc.Response {
		order = append(order, "handler")
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}, handlerMw)

	mc := &MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	r.Dispatch(mc)

	assert.Equal(t, []string{"global", "handler-mw", "handler"}, order)
}

func TestRouter_Methods(t *testing.T) {
	r := NewRouter()

	r.Handle("method.a", func(mc *MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "a")
	})
	r.Handle("method.b", func(mc *MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "b")
	})

	methods := r.Methods()

	assert.Len(t, methods, 2)
	assert.Contains(t, methods, "method.a")
	assert.Contains(t, methods, "method.b")
}
