// ABOUTME: Tests for middleware types and chain composition.
// ABOUTME: Verifies MiddlewareContext and middleware chaining behavior.

package ws

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"bingo/internal/pkg/contextx"
	"bingo/pkg/jsonrpc"
)

func TestMiddlewareContext_RequestID(t *testing.T) {
	ctx := context.Background()
	ctx = contextx.WithRequestID(ctx, "test-123")

	mc := &MiddlewareContext{
		Ctx:       ctx,
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Method:    "test",
		StartTime: time.Now(),
	}

	assert.Equal(t, "test-123", mc.RequestID())
}

func TestMiddlewareContext_UserID(t *testing.T) {
	ctx := context.Background()
	ctx = contextx.WithUserID(ctx, "user-456")

	mc := &MiddlewareContext{
		Ctx:       ctx,
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Method:    "test",
		StartTime: time.Now(),
	}

	assert.Equal(t, "user-456", mc.UserID())
}

func TestMiddlewareContext_EmptyContext(t *testing.T) {
	mc := &MiddlewareContext{
		Ctx:     nil,
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	assert.Equal(t, "", mc.RequestID())
	assert.Equal(t, "", mc.UserID())
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

func TestMiddlewareChain_Empty(t *testing.T) {
	handler := func(mc *MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	chain := Chain()
	wrapped := chain(handler)

	mc := &MiddlewareContext{
		Request: &jsonrpc.Request{ID: 1},
	}
	resp := wrapped(mc)

	assert.Equal(t, "ok", resp.Result)
}
