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
	return func(mc *ws.Context) *jsonrpc.Response {
		if mc.Client == nil || !mc.Client.IsAuthenticated() {
			return jsonrpc.NewErrorResponse(mc.Request.ID,
				errorsx.New(401, "Unauthorized", "Login required"))
		}

		// Add user ID to context
		mc.Ctx = contextx.WithUserID(mc.Ctx, mc.Client.UserID)

		return next(mc)
	}
}
