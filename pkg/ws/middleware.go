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

// MiddlewareContext contains all information needed by middleware.
type MiddlewareContext struct {
	Ctx       context.Context
	Request   *jsonrpc.Request
	Client    *Client
	Method    string
	StartTime time.Time
}

// RequestID returns the request ID from context.
func (mc *MiddlewareContext) RequestID() string {
	if mc.Ctx == nil {
		return ""
	}
	return contextx.RequestID(mc.Ctx)
}

// UserID returns the user ID from context.
func (mc *MiddlewareContext) UserID() string {
	if mc.Ctx == nil {
		return ""
	}
	return contextx.UserID(mc.Ctx)
}

// BindParams unmarshals the request params into the given struct.
func (mc *MiddlewareContext) BindParams(v any) error {
	if len(mc.Request.Params) == 0 {
		return nil
	}
	return json.Unmarshal(mc.Request.Params, v)
}

// BindValidate unmarshals and validates the request params.
func (mc *MiddlewareContext) BindValidate(v any) error {
	if err := mc.BindParams(v); err != nil {
		return err
	}
	return validate.Struct(v)
}

// Handler is a message handler function.
type Handler func(*MiddlewareContext) *jsonrpc.Response

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
