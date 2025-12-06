// ABOUTME: Request logging middleware for WebSocket handlers.
// ABOUTME: Logs method, latency, and error status.

package middleware

import (
	"time"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// LoggerWithLogger logs request details after handling using the provided logger.
func LoggerWithLogger(logger ws.Logger) ws.Middleware {
	return func(next ws.Handler) ws.Handler {
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
				logger.WithContext(c.Context).Warnw("WebSocket request failed", fields...)
			} else {
				logger.WithContext(c.Context).Infow("WebSocket request", fields...)
			}

			return resp
		}
	}
}

// Logger logs request details after handling.
var Logger = LoggerWithLogger(ws.DefaultLogger())
