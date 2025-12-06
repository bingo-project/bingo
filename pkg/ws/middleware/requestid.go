// ABOUTME: Request ID middleware for WebSocket handlers.
// ABOUTME: Uses client-provided ID or generates UUID.

package middleware

import (
	"fmt"

	"github.com/google/uuid"

	"bingo/internal/pkg/contextx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// RequestID adds request ID to context.
// Uses client-provided ID if present, otherwise generates UUID.
func RequestID(next ws.Handler) ws.Handler {
	return func(mc *ws.Context) *jsonrpc.Response {
		requestID := ""
		if mc.Request.ID != nil {
			requestID = fmt.Sprintf("%v", mc.Request.ID)
		}
		if requestID == "" {
			requestID = uuid.New().String()
		}

		mc.Ctx = contextx.WithRequestID(mc.Ctx, requestID)
		return next(mc)
	}
}
