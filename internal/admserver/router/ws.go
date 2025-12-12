// ABOUTME: WebSocket method registration for JSON-RPC.
// ABOUTME: Maps JSON-RPC methods to handler layer methods via Router.

package router

import (
	"github.com/bingo-project/websocket"
	"github.com/bingo-project/websocket/middleware"
)

// RegisterWSHandlers registers all WebSocket handlers with the router.
func RegisterWSHandlers(router *websocket.Router, rateLimitStore *middleware.RateLimiterStore, logger websocket.Logger) {
	// Global middleware
	router.Use(
		middleware.RecoveryWithLogger(logger),
		middleware.RequestID,
		middleware.LoggerWithLogger(logger),
		middleware.RateLimitWithStore(&middleware.RateLimitConfig{
			Default: 10,
			Methods: map[string]float64{
				"heartbeat": 0, // No limit for heartbeat
			},
		}, rateLimitStore),
	)

	// Public methods (no auth required)
	public := router.Group()
	public.Handle("heartbeat", websocket.HeartbeatHandler)

	// Private methods (require auth)
	// private := router.Group(middleware.Auth)
	// Admin methods can be registered here as needed.
}
