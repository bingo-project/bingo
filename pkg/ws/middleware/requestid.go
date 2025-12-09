// ABOUTME: Request ID middleware for WebSocket handlers.
// ABOUTME: Uses client-provided ID or generates UUID.

package middleware

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/bingo-project/bingo/pkg/contextx"
	"github.com/bingo-project/bingo/pkg/jsonrpc"
	"github.com/bingo-project/bingo/pkg/ws"
)

// RequestID adds request ID to context.
// Uses client-provided ID if present, otherwise generates UUID.
func RequestID(next ws.Handler) ws.Handler {
	return func(c *ws.Context) *jsonrpc.Response {
		requestID := ""
		if c.Request.ID != nil {
			requestID = fmt.Sprintf("%v", c.Request.ID)
		}
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Context = contextx.WithRequestID(c.Context, requestID)

		return next(c)
	}
}
