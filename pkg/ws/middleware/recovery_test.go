// ABOUTME: Tests for panic recovery middleware.
// ABOUTME: Verifies panics are caught and converted to error responses.

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

func TestRecovery_PanicWithError(t *testing.T) {
	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		panic(assert.AnError)
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
}
