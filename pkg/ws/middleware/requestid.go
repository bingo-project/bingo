// ABOUTME: Request ID middleware for WebSocket handlers.
// ABOUTME: Uses client-provided ID or generates UUID.

package middleware

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// RequestID adds request ID to context.
// Uses client-provided ID if present, otherwise generates UUID.
func RequestID(next ws.Handler) ws.Handler {
	return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		requestID := ""
		if mc.Request.ID != nil {
			requestID = fmt.Sprintf("%v", mc.Request.ID)
		}
		if requestID == "" {
			requestID = uuid.New().String()
		}

		mc.Ctx = context.WithValue(mc.Ctx, ws.RequestIDKey{}, requestID)
		return next(mc)
	}
}
