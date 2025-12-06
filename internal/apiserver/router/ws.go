// ABOUTME: WebSocket method registration for JSON-RPC.
// ABOUTME: Maps JSON-RPC methods to handler layer methods via Router.

package router

import (
	"context"
	"encoding/json"

	"bingo/internal/apiserver/biz"
	wshandler "bingo/internal/apiserver/handler/ws"
	"bingo/pkg/api/apiserver/v1"
	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
	"bingo/pkg/ws/middleware"
)

// RegisterWSHandlers registers all WebSocket handlers with the router.
func RegisterWSHandlers(router *ws.Router, b biz.IBiz) {
	systemHandler := wshandler.NewSystemHandler(b)
	authHandler := wshandler.NewAuthHandler(b)

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
	public.Handle("system.healthz", wrapBizHandler(systemHandler.Healthz))
	public.Handle("system.version", wrapBizHandler(systemHandler.Version))
	public.Handle("auth.login", wrapBizHandlerWithReq(authHandler.Login, &v1.LoginRequest{}))

	// Private methods (require auth)
	private := router.Group(middleware.Auth)
	private.Handle("subscribe", ws.SubscribeHandler)
	private.Handle("unsubscribe", ws.UnsubscribeHandler)
	private.Handle("auth.user-info", wrapBizHandler(authHandler.UserInfo))
}

// wrapBizHandler adapts a biz handler (no request params) to ws.Handler.
func wrapBizHandler(handler func(context.Context, any) (any, error)) ws.Handler {
	return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		resp, err := handler(mc.Ctx, nil)
		if err != nil {
			return jsonrpc.NewErrorResponse(mc.Request.ID, err)
		}
		return jsonrpc.NewResponse(mc.Request.ID, resp)
	}
}

// wrapBizHandlerWithReq adapts a biz handler with request params to ws.Handler.
func wrapBizHandlerWithReq[T any](handler func(context.Context, any) (any, error), _ *T) ws.Handler {
	return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		req := new(T)
		if len(mc.Request.Params) > 0 {
			if err := json.Unmarshal(mc.Request.Params, req); err != nil {
				return jsonrpc.NewErrorResponse(mc.Request.ID,
					errorsx.New(400, "InvalidParams", "Invalid params: %s", err.Error()))
			}
		}

		resp, err := handler(mc.Ctx, req)
		if err != nil {
			return jsonrpc.NewErrorResponse(mc.Request.ID, err)
		}
		return jsonrpc.NewResponse(mc.Request.ID, resp)
	}
}
