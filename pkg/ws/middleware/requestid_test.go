// ABOUTME: Tests for request ID middleware.
// ABOUTME: Verifies request ID is added to context from client ID or generated.

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

	handler := func(c *ws.Context) *jsonrpc.Response {
		capturedRequestID = c.RequestID()

		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := RequestID(handler)

	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: "client-123", Method: "test"},
		Method:  "test",
	}

	wrapped(c)

	assert.Equal(t, "client-123", capturedRequestID)
}

func TestRequestID_GeneratesIfMissing(t *testing.T) {
	var capturedRequestID string

	handler := func(c *ws.Context) *jsonrpc.Response {
		capturedRequestID = c.RequestID()

		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := RequestID(handler)

	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{Method: "test"}, // No ID
		Method:  "test",
	}

	wrapped(c)

	assert.NotEmpty(t, capturedRequestID)
	assert.Len(t, capturedRequestID, 36) // UUID length
}

func TestRequestID_NumericID(t *testing.T) {
	var capturedRequestID string

	handler := func(c *ws.Context) *jsonrpc.Response {
		capturedRequestID = c.RequestID()

		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := RequestID(handler)

	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 42, Method: "test"},
		Method:  "test",
	}

	wrapped(c)

	assert.Equal(t, "42", capturedRequestID)
}
