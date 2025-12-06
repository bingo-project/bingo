// ABOUTME: Authentication middleware for WebSocket handlers.
// ABOUTME: Blocks unauthenticated requests with 401 error.

package middleware

import (
	"bingo/internal/pkg/contextx"
	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// Auth requires the client to be authenticated.
func Auth(next ws.Handler) ws.Handler {
	return func(c *ws.Context) *jsonrpc.Response {
		if c.Client == nil || !c.Client.IsAuthenticated() {
			return jsonrpc.NewErrorResponse(c.Request.ID,
				errorsx.New(401, "Unauthorized", "Login required"))
		}

		// Add user ID to context
		c.Context = contextx.WithUserID(c.Context, c.Client.UserID)

		return next(c)
	}
}
