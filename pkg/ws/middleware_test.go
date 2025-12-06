// ABOUTME: Tests for middleware types and chain composition.
// ABOUTME: Verifies Context and middleware chaining behavior.

package ws

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"bingo/internal/pkg/contextx"
	"bingo/pkg/jsonrpc"
)

func TestContext_RequestID(t *testing.T) {
	ctx := context.Background()
	ctx = contextx.WithRequestID(ctx, "test-123")

	c := &Context{
		Ctx:       ctx,
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Method:    "test",
		StartTime: time.Now(),
	}

	assert.Equal(t, "test-123", c.RequestID())
}

func TestContext_UserID(t *testing.T) {
	ctx := context.Background()
	ctx = contextx.WithUserID(ctx, "user-456")

	c := &Context{
		Ctx:       ctx,
		Request:   &jsonrpc.Request{ID: 1, Method: "test"},
		Method:    "test",
		StartTime: time.Now(),
	}

	assert.Equal(t, "user-456", c.UserID())
}

func TestContext_EmptyContext(t *testing.T) {
	c := &Context{
		Ctx:     nil,
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Method:  "test",
	}

	assert.Equal(t, "", c.RequestID())
	assert.Equal(t, "", c.UserID())
}

func TestMiddlewareChain(t *testing.T) {
	var order []string

	m1 := func(next Handler) Handler {
		return func(c *Context) *jsonrpc.Response {
			order = append(order, "m1-before")
			resp := next(c)
			order = append(order, "m1-after")
			return resp
		}
	}

	m2 := func(next Handler) Handler {
		return func(c *Context) *jsonrpc.Response {
			order = append(order, "m2-before")
			resp := next(c)
			order = append(order, "m2-after")
			return resp
		}
	}

	handler := func(c *Context) *jsonrpc.Response {
		order = append(order, "handler")
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	chain := Chain(m1, m2)
	wrapped := chain(handler)

	c := &Context{
		Request: &jsonrpc.Request{ID: 1},
	}
	wrapped(c)

	assert.Equal(t, []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}, order)
}

func TestMiddlewareChain_Empty(t *testing.T) {
	handler := func(c *Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	chain := Chain()
	wrapped := chain(handler)

	c := &Context{
		Request: &jsonrpc.Request{ID: 1},
	}
	resp := wrapped(c)

	assert.Equal(t, "ok", resp.Result)
}

func TestContext_BindParams(t *testing.T) {
	c := &Context{
		Request: &jsonrpc.Request{
			ID:     1,
			Method: "test",
			Params: []byte(`{"username":"alice","age":25}`),
		},
	}

	var params struct {
		Username string `json:"username"`
		Age      int    `json:"age"`
	}

	err := c.BindParams(&params)
	assert.NoError(t, err)
	assert.Equal(t, "alice", params.Username)
	assert.Equal(t, 25, params.Age)
}

func TestContext_BindParams_Empty(t *testing.T) {
	c := &Context{
		Request: &jsonrpc.Request{
			ID:     1,
			Method: "test",
			Params: nil,
		},
	}

	var params struct {
		Username string `json:"username"`
	}

	err := c.BindParams(&params)
	assert.NoError(t, err)
	assert.Equal(t, "", params.Username)
}

func TestContext_BindParams_Invalid(t *testing.T) {
	c := &Context{
		Request: &jsonrpc.Request{
			ID:     1,
			Method: "test",
			Params: []byte(`invalid json`),
		},
	}

	var params struct {
		Username string `json:"username"`
	}

	err := c.BindParams(&params)
	assert.Error(t, err)
}

func TestContext_BindValidate(t *testing.T) {
	c := &Context{
		Request: &jsonrpc.Request{
			ID:     1,
			Method: "test",
			Params: []byte(`{"username":"alice","email":"alice@example.com"}`),
		},
	}

	var params struct {
		Username string `json:"username" validate:"required,min=3"`
		Email    string `json:"email" validate:"required,email"`
	}

	err := c.BindValidate(&params)
	assert.NoError(t, err)
	assert.Equal(t, "alice", params.Username)
	assert.Equal(t, "alice@example.com", params.Email)
}

func TestContext_BindValidate_ValidationError(t *testing.T) {
	c := &Context{
		Request: &jsonrpc.Request{
			ID:     1,
			Method: "test",
			Params: []byte(`{"username":"ab","email":"invalid"}`),
		},
	}

	var params struct {
		Username string `json:"username" validate:"required,min=3"`
		Email    string `json:"email" validate:"required,email"`
	}

	err := c.BindValidate(&params)
	assert.Error(t, err)
}
