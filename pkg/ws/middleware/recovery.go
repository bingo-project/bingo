// ABOUTME: Panic recovery middleware for WebSocket handlers.
// ABOUTME: Catches panics and returns a JSON-RPC error response.

package middleware

import (
	"runtime/debug"

	"github.com/bingo-project/bingo/pkg/errorsx"
	"github.com/bingo-project/bingo/pkg/jsonrpc"
	"github.com/bingo-project/bingo/pkg/ws"
)

// RecoveryWithLogger catches panics and returns an error response using the provided logger.
func RecoveryWithLogger(logger ws.Logger) ws.Middleware {
	return func(next ws.Handler) ws.Handler {
		return func(c *ws.Context) (resp *jsonrpc.Response) {
			defer func() {
				if r := recover(); r != nil {
					logger.WithContext(c.Context).Errorw("WebSocket panic recovered",
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
}

// Recovery catches panics and returns an error response.
// Uses NopLogger by default; use RecoveryWithLogger to inject a real logger.
var Recovery = RecoveryWithLogger(ws.NopLogger())
