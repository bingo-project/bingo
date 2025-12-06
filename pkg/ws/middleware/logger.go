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
	return func(c *ws.Context) *jsonrpc.Response {
		resp := next(c)

		fields := []any{
			"method", c.Method,
			"latency", time.Since(c.StartTime),
		}

		if c.Client != nil {
			fields = append(fields, "client_id", c.Client.ID, "client_addr", c.Client.Addr)
			if c.Client.UserID != "" {
				fields = append(fields, "user_id", c.Client.UserID)
			}
		}

		if resp.Error != nil {
			fields = append(fields, "error", resp.Error.Reason)
			log.C(c.Ctx).Warnw("WebSocket request failed", fields...)
		} else {
			log.C(c.Ctx).Infow("WebSocket request", fields...)
		}

		return resp
	}
}
