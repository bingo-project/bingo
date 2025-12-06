// ABOUTME: Middleware types for WebSocket message handling.
// ABOUTME: Provides middleware chain composition similar to HTTP/gRPC patterns.

package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"

	"bingo/internal/pkg/contextx"
	"bingo/pkg/jsonrpc"
)

// validate is the singleton validator instance.
var validate = validator.New()

// Context contains all information needed by middleware.
type Context struct {
	Ctx       context.Context
	Request   *jsonrpc.Request
	Client    *Client
	Method    string
	StartTime time.Time
}

// RequestID returns the request ID from context.
func (c *Context) RequestID() string {
	if c.Ctx == nil {
		return ""
	}
	return contextx.RequestID(c.Ctx)
}

// UserID returns the user ID from context.
func (c *Context) UserID() string {
	if c.Ctx == nil {
		return ""
	}
	return contextx.UserID(c.Ctx)
}

// BindParams unmarshals the request params into the given struct.
func (c *Context) BindParams(v any) error {
	if len(c.Request.Params) == 0 {
		return nil
	}
	return json.Unmarshal(c.Request.Params, v)
}

// BindValidate unmarshals and validates the request params.
func (c *Context) BindValidate(v any) error {
	if err := c.BindParams(v); err != nil {
		return err
	}
	return validate.Struct(v)
}

// Handler is a message handler function.
type Handler func(*Context) *jsonrpc.Response

// Middleware wraps a handler with additional functionality.
type Middleware func(Handler) Handler

// Chain combines multiple middlewares into a single middleware.
func Chain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
