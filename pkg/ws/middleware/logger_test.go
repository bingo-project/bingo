// ABOUTME: Tests for logger middleware.
// ABOUTME: Verifies request logging for success and error responses.

package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

func TestLogger_Success(t *testing.T) {
	handler := func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := Logger(handler)

	c := &ws.Context{
		Ctx:       context.Background(),
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Method:    "test",
		StartTime: time.Now(),
	}

	resp := wrapped(c)

	assert.Nil(t, resp.Error)
	// Logger should not modify the response
	assert.Equal(t, "ok", resp.Result)
}

func TestLogger_Error(t *testing.T) {
	handler := func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewErrorResponse(c.Request.ID,
			errorsx.New(400, "BadRequest", "test error"))
	}

	wrapped := Logger(handler)

	c := &ws.Context{
		Ctx:       context.Background(),
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Method:    "test",
		StartTime: time.Now(),
	}

	resp := wrapped(c)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, "BadRequest", resp.Error.Reason)
}

func TestLogger_WithClient(t *testing.T) {
	handler := func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := Logger(handler)

	client := &ws.Client{
		Addr:   "127.0.0.1:12345",
		UserID: "user-123",
	}

	c := &ws.Context{
		Ctx:       context.Background(),
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Client:    client,
		Method:    "test",
		StartTime: time.Now(),
	}

	resp := wrapped(c)

	assert.Nil(t, resp.Error)
	assert.Equal(t, "ok", resp.Result)
}
