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
	public.Handle("auth.login", wrapLoginHandler(authHandler.Login))

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

// loginParams extends LoginRequest with WebSocket-specific fields.
type loginParams struct {
	v1.LoginRequest
	Platform string `json:"platform"`
}

// wrapLoginHandler adapts the login handler to update client state after successful login.
func wrapLoginHandler(handler func(context.Context, any) (any, error)) ws.Handler {
	return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
		// Parse request params
		var params loginParams
		if len(mc.Request.Params) > 0 {
			if err := json.Unmarshal(mc.Request.Params, &params); err != nil {
				return jsonrpc.NewErrorResponse(mc.Request.ID,
					errorsx.New(400, "InvalidParams", "Invalid params: %s", err.Error()))
			}
		}

		// Validate platform
		if !ws.IsValidPlatform(params.Platform) {
			return jsonrpc.NewErrorResponse(mc.Request.ID,
				errorsx.New(400, "InvalidPlatform", "Invalid platform: %s", params.Platform))
		}

		// Call login handler
		resp, err := handler(mc.Ctx, &params.LoginRequest)
		if err != nil {
			return jsonrpc.NewErrorResponse(mc.Request.ID, err)
		}

		// Parse token from response to get user info
		if mc.Client != nil {
			respBytes, _ := json.Marshal(resp)
			var loginResp struct {
				AccessToken string `json:"accessToken"`
			}
			if err := json.Unmarshal(respBytes, &loginResp); err == nil && loginResp.AccessToken != "" {
				// Parse token using client's token parser
				tokenInfo, err := mc.Client.ParseToken(loginResp.AccessToken)
				if err == nil {
					// Update client context
					mc.Client.UpdateContext(tokenInfo.UserID)

					// Notify hub about login
					mc.Client.NotifyLogin(tokenInfo.UserID, params.Platform, tokenInfo.ExpiresAt)
				}
			}
		}

		return jsonrpc.NewResponse(mc.Request.ID, resp)
	}
}
