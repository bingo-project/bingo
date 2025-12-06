// ABOUTME: Middleware types for WebSocket message handling.
// ABOUTME: Provides middleware chain composition similar to HTTP/gRPC patterns.

package ws

import (
	"context"
	"time"

	"bingo/internal/pkg/contextx"
	"bingo/pkg/jsonrpc"
)

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
