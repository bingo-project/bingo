// ABOUTME: Panic recovery middleware for WebSocket handlers.
// ABOUTME: Catches panics and returns a JSON-RPC error response.

package middleware

import (
	"runtime/debug"

	"github.com/bingo-project/component-base/log"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// Recovery catches panics and returns an error response.
func Recovery(next ws.Handler) ws.Handler {
	return func(c *ws.Context) (resp *jsonrpc.Response) {
		defer func() {
			if r := recover(); r != nil {
				log.C(c.Ctx).Errorw("WebSocket panic recovered",
					"method", c.Method,
					"panic", r,
					"stack", string(debug.Stack()),
				)
				resp = jsonrpc.NewErrorResponse(c.Request.ID,
					errorsx.New(500, "InternalError", "Internal server error"))
			}
		}()
		return next(c)
	}
}
