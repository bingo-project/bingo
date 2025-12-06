// ABOUTME: Login middleware for WebSocket handlers.
// ABOUTME: Updates client authentication state after successful login.

package middleware

import (
	"encoding/json"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// loginParams contains the platform field from login request.
type loginParams struct {
	Platform string `json:"platform"`
}

// loginResponse contains the accessToken field from login response.
type loginResponse struct {
	AccessToken string `json:"accessToken"`
}

// LoginStateUpdater updates client state after successful login.
// It validates platform, and after successful login, parses the access token
// from response and notifies the hub.
func LoginStateUpdater(next ws.Handler) ws.Handler {
	return func(mc *ws.Context) *jsonrpc.Response {
		// Parse platform from request params
		var params loginParams
		if len(mc.Request.Params) > 0 {
			_ = json.Unmarshal(mc.Request.Params, &params)
		}

		// Validate platform
		if !ws.IsValidPlatform(params.Platform) {
			return jsonrpc.NewErrorResponse(mc.Request.ID,
				errorsx.New(400, "InvalidPlatform", "Invalid platform: %s", params.Platform))
		}

		// Call next handler
		resp := next(mc)

		// Only process successful responses
		if resp.Error != nil || mc.Client == nil {
			return resp
		}

		// Parse token from response
		respBytes, _ := json.Marshal(resp.Result)
		var loginResp loginResponse
		if err := json.Unmarshal(respBytes, &loginResp); err != nil || loginResp.AccessToken == "" {
			return resp
		}

		// Parse token using client's token parser
		tokenInfo, err := mc.Client.ParseToken(loginResp.AccessToken)
		if err != nil {
			return resp
		}

		// Update client context
		mc.Client.UpdateContext(tokenInfo.UserID)

		// Notify hub about login
		mc.Client.NotifyLogin(tokenInfo.UserID, params.Platform, tokenInfo.ExpiresAt)

		return resp
	}
}
