// ABOUTME: Tests for authentication middleware.
// ABOUTME: Verifies authenticated clients pass and unauthenticated are rejected.

package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bingo-project/bingo/pkg/jsonrpc"
	"github.com/bingo-project/bingo/pkg/ws"
)

func TestAuth_Authenticated(t *testing.T) {
	handler := func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := Auth(handler)

	client := &ws.Client{}
	client.UserID = "user-123"
	client.Platform = "web"
	client.LoginTime = 1000

	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  client,
		Method:  "test",
	}

	resp := wrapped(c)

	assert.Nil(t, resp.Error)
	assert.Equal(t, "ok", resp.Result)
}

func TestAuth_Unauthenticated(t *testing.T) {
	handler := func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := Auth(handler)

	client := &ws.Client{} // Not logged in

	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  client,
		Method:  "test",
	}

	resp := wrapped(c)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, "Unauthorized", resp.Error.Reason)
}

func TestAuth_NilClient(t *testing.T) {
	handler := func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := Auth(handler)

	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  nil,
		Method:  "test",
	}

	resp := wrapped(c)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, "Unauthorized", resp.Error.Reason)
}

func TestAuth_SetsUserIDInContext(t *testing.T) {
	var capturedUserID string

	handler := func(c *ws.Context) *jsonrpc.Response {
		capturedUserID = c.UserID()

		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := Auth(handler)

	client := &ws.Client{}
	client.UserID = "user-456"
	client.Platform = "web"
	client.LoginTime = 1000

	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  client,
		Method:  "test",
	}

	wrapped(c)

	assert.Equal(t, "user-456", capturedUserID)
}
