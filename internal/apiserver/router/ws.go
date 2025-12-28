// ABOUTME: WebSocket method registration for JSON-RPC.
// ABOUTME: Maps JSON-RPC methods to handler layer methods via Router.

package router

import (
	"github.com/bingo-project/websocket"
	"github.com/bingo-project/websocket/middleware"

	wshandler "github.com/bingo-project/bingo/internal/apiserver/handler/ws"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

// RegisterWSHandlers registers all WebSocket handlers with the router.
func RegisterWSHandlers(router *websocket.Router, rateLimitStore *middleware.RateLimiterStore, logger websocket.Logger) {
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
	public.Handle("heartbeat", websocket.HeartbeatHandler)
	public.Handle("system.healthz", h.Healthz)
	public.Handle("system.version", h.Version)
	// LoginStateUpdater runs after login handler succeeds to update client state.
	public.Handle("auth.login", h.Login, middleware.LoginStateUpdater)
	public.Handle("auth.loginByToken", websocket.TokenLoginHandler, middleware.LoginStateUpdater)

	// Private methods (require auth)
	private := router.Group(middleware.Auth)
	private.Handle("subscribe", websocket.SubscribeHandler)
	private.Handle("unsubscribe", websocket.UnsubscribeHandler)
	private.Handle("auth.user-info", h.UserInfo)
}
