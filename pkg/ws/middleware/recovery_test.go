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
	handler := func(c *ws.Context) *jsonrpc.Response {
		panic("test panic")
	}

	wrapped := Recovery(handler)

	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	resp := wrapped(c)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, "InternalError", resp.Error.Reason)
	assert.Equal(t, "Internal server error", resp.Error.Message)
	assert.NotContains(t, resp.Error.Message, "panic") // Panic details should not be exposed
}

func TestRecovery_NoError(t *testing.T) {
	handler := func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := Recovery(handler)

	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	resp := wrapped(c)

	assert.Nil(t, resp.Error)
	assert.Equal(t, "ok", resp.Result)
}

func TestRecovery_PanicWithError(t *testing.T) {
	handler := func(c *ws.Context) *jsonrpc.Response {
		panic(assert.AnError)
	}

	wrapped := Recovery(handler)

	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	resp := wrapped(c)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, "InternalError", resp.Error.Reason)
}
