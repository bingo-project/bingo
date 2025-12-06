// ABOUTME: WebSocket method registration for JSON-RPC.
// ABOUTME: Maps JSON-RPC methods to handler layer methods via Router.

package router

import (
	"bingo/pkg/ws"
	"bingo/pkg/ws/middleware"
)

// RegisterWSHandlers registers all WebSocket handlers with the router.
func RegisterWSHandlers(router *ws.Router) {
	// Global middleware
	router.Use(
		middleware.Recovery,
		middleware.RequestID,
		middleware.Logger,
		middleware.RateLimit(&middleware.RateLimitConfig{
			Default: 10,
			Methods: map[string]float64{
				"heartbeat": 0, // No limit for heartbeat
			},
		}),
	)

	// Public methods (no auth required)
	public := router.Group()
	public.Handle("heartbeat", ws.HeartbeatHandler)

	// Private methods (require auth)
	// private := router.Group(middleware.Auth)
	// Admin methods can be registered here as needed.
}
