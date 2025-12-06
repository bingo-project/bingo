// ABOUTME: Built-in handlers for common WebSocket methods.
// ABOUTME: Provides heartbeat, subscribe, unsubscribe, and handler adapters.

package ws

import (
	"context"
	"encoding/json"
	"time"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
)

// BizHandler is a business logic handler function type.
type BizHandler func(ctx context.Context, req any) (any, error)

// WrapBizHandler adapts a biz handler (no request params) to ws.Handler.
func WrapBizHandler(handler BizHandler) Handler {
	return func(mc *MiddlewareContext) *jsonrpc.Response {
		resp, err := handler(mc.Ctx, nil)
		if err != nil {
			return jsonrpc.NewErrorResponse(mc.Request.ID, err)
		}
		return jsonrpc.NewResponse(mc.Request.ID, resp)
	}
}

// WrapParamsHandler adapts a biz handler with typed request params to ws.Handler.
// The paramsFactory creates a new instance of the params type for unmarshaling.
func WrapParamsHandler[T any](handler BizHandler, paramsFactory func() *T) Handler {
	return func(mc *MiddlewareContext) *jsonrpc.Response {
		params := paramsFactory()
		if len(mc.Request.Params) > 0 {
			if err := json.Unmarshal(mc.Request.Params, params); err != nil {
				return jsonrpc.NewErrorResponse(mc.Request.ID,
					errorsx.New(400, "InvalidParams", "Invalid params: %s", err.Error()))
			}
		}

		resp, err := handler(mc.Ctx, params)
		if err != nil {
			return jsonrpc.NewErrorResponse(mc.Request.ID, err)
		}
		return jsonrpc.NewResponse(mc.Request.ID, resp)
	}
}

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
