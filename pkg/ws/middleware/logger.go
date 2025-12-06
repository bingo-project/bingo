// ABOUTME: Request logging middleware for WebSocket handlers.
// ABOUTME: Logs method, latency, and error status.

package middleware

import (
	"time"

	"github.com/bingo-project/component-base/log"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// Logger logs request details after handling.
func Logger(next ws.Handler) ws.Handler {
	return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		resp := next(mc)

		fields := []any{
			"method", mc.Method,
			"latency", time.Since(mc.StartTime),
		}

		if mc.Client != nil {
			fields = append(fields, "client_id", mc.Client.ID, "client_addr", mc.Client.Addr)
			if mc.Client.UserID != "" {
				fields = append(fields, "user_id", mc.Client.UserID)
			}
		}

		if resp.Error != nil {
			fields = append(fields, "error", resp.Error.Reason)
			log.C(mc.Ctx).Warnw("WebSocket request failed", fields...)
		} else {
			log.C(mc.Ctx).Infow("WebSocket request", fields...)
		}

		return resp
	}
}
