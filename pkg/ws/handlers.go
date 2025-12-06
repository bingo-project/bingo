// ABOUTME: Built-in handlers for common WebSocket methods.
// ABOUTME: Provides heartbeat, subscribe, and unsubscribe functionality.

package ws

import (
	"encoding/json"
	"time"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
)

// HeartbeatHandler responds to heartbeat requests.
func HeartbeatHandler(mc *MiddlewareContext) *jsonrpc.Response {
	return jsonrpc.NewResponse(mc.Request.ID, map[string]any{
		"status":      "ok",
		"server_time": time.Now().Unix(),
	})
}

// SubscribeHandler handles topic subscription.
func SubscribeHandler(mc *MiddlewareContext) *jsonrpc.Response {
	var params struct {
		Topics []string `json:"topics"`
	}

	if err := json.Unmarshal(mc.Request.Params, &params); err != nil {
		return jsonrpc.NewErrorResponse(mc.Request.ID,
			errorsx.New(400, "InvalidParams", "Invalid subscribe params"))
	}

	if len(params.Topics) == 0 {
		return jsonrpc.NewErrorResponse(mc.Request.ID,
			errorsx.New(400, "InvalidParams", "Topics required"))
	}

	result := make(chan []string, 1)
	mc.Client.hub.Subscribe <- &SubscribeEvent{
		Client: mc.Client,
		Topics: params.Topics,
		Result: result,
	}

	subscribed := <-result
	return jsonrpc.NewResponse(mc.Request.ID, map[string]any{
		"subscribed": subscribed,
	})
}

// UnsubscribeHandler handles topic unsubscription.
func UnsubscribeHandler(mc *MiddlewareContext) *jsonrpc.Response {
	var params struct {
		Topics []string `json:"topics"`
	}

	if err := json.Unmarshal(mc.Request.Params, &params); err != nil {
		return jsonrpc.NewErrorResponse(mc.Request.ID,
			errorsx.New(400, "InvalidParams", "Invalid unsubscribe params"))
	}

	if len(params.Topics) == 0 {
		return jsonrpc.NewErrorResponse(mc.Request.ID,
			errorsx.New(400, "InvalidParams", "Topics required"))
	}

	mc.Client.hub.Unsubscribe <- &UnsubscribeEvent{
		Client: mc.Client,
		Topics: params.Topics,
	}

	return jsonrpc.NewResponse(mc.Request.ID, map[string]any{
		"unsubscribed": params.Topics,
	})
}
