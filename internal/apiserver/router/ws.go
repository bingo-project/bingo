// ABOUTME: WebSocket method registration for JSON-RPC.
// ABOUTME: Maps JSON-RPC methods to handler layer methods via Router.

package router

import (
	wshandler "bingo/internal/apiserver/handler/ws"
	"bingo/internal/pkg/store"
	"bingo/pkg/api/apiserver/v1"
	"bingo/pkg/ws"
	"bingo/pkg/ws/middleware"
)

// RegisterWSHandlers registers all WebSocket handlers with the router.
func RegisterWSHandlers(router *ws.Router) {
	h := wshandler.NewHandler(store.S)

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
	public.Handle("system.healthz", ws.WrapBizHandler(h.Healthz))
	public.Handle("system.version", ws.WrapBizHandler(h.Version))
	public.Handle("auth.login", middleware.LoginStateUpdater(
		ws.WrapParamsHandler(h.Login, func() *v1.LoginRequest { return &v1.LoginRequest{} }),
	))

	// Private methods (require auth)
	private := router.Group(middleware.Auth)
	private.Handle("subscribe", ws.SubscribeHandler)
	private.Handle("unsubscribe", ws.UnsubscribeHandler)
	private.Handle("auth.user-info", ws.WrapBizHandler(h.UserInfo))
}
