// ABOUTME: Tests for authentication middleware.
// ABOUTME: Verifies authenticated clients pass and unauthenticated are rejected.

package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

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

func TestAuth_NilClient(t *testing.T) {
	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	wrapped := Auth(handler)

	mc := &ws.MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  nil,
		Method:  "test",
	}

	resp := wrapped(mc)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, "Unauthorized", resp.Error.Reason)
}

func TestAuth_SetsUserIDInContext(t *testing.T) {
	var capturedUserID string

	handler := func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		capturedUserID = mc.UserID()
		return jsonrpc.NewResponse(mc.Request.ID, "ok")
	}

	wrapped := Auth(handler)

	client := &ws.Client{}
	client.UserID = "user-456"
	client.Platform = "web"
	client.LoginTime = 1000

	mc := &ws.MiddlewareContext{
		Ctx:     context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  client,
		Method:  "test",
	}

	wrapped(mc)

	assert.Equal(t, "user-456", capturedUserID)
}
