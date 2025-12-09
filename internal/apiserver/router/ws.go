// ABOUTME: WebSocket method registration for JSON-RPC.
// ABOUTME: Maps JSON-RPC methods to handler layer methods via Router.

package router

import (
	wshandler "github.com/bingo-project/bingo/internal/apiserver/handler/ws"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ws"
	"github.com/bingo-project/bingo/pkg/ws/middleware"
)

// RegisterWSHandlers registers all WebSocket handlers with the router.
func RegisterWSHandlers(router *ws.Router, rateLimitStore *middleware.RateLimiterStore, logger ws.Logger) {
	h := wshandler.NewHandler(store.S)

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
	public.Handle("heartbeat", ws.HeartbeatHandler)
	public.Handle("system.healthz", h.Healthz)
	public.Handle("system.version", h.Version)
	// LoginStateUpdater runs after Login handler succeeds, parsing the returned
	// access token to update client state and notify the hub of the login event.
	public.Handle("auth.login", h.Login, middleware.LoginStateUpdater)

	// Private methods (require auth)
	private := router.Group(middleware.Auth)
	private.Handle("subscribe", ws.SubscribeHandler)
	private.Handle("unsubscribe", ws.UnsubscribeHandler)
	private.Handle("auth.user-info", h.UserInfo)
}
