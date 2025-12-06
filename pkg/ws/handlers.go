// ABOUTME: Built-in handlers for common WebSocket methods.
// ABOUTME: Provides heartbeat, subscribe, and unsubscribe handlers.

package ws

import (
	"time"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
)

// HeartbeatHandler responds to heartbeat requests.
func HeartbeatHandler(c *Context) *jsonrpc.Response {
	return jsonrpc.NewResponse(c.Request.ID, map[string]any{
		"status":      "ok",
		"server_time": time.Now().Unix(),
	})
}

// SubscribeHandler handles topic subscription.
func SubscribeHandler(c *Context) *jsonrpc.Response {
	var params struct {
		Topics []string `json:"topics"`
	}

	if err := c.BindParams(&params); err != nil {
		return jsonrpc.NewErrorResponse(c.Request.ID,
			errorsx.New(400, "InvalidParams", "Invalid subscribe params"))
	}

	if len(params.Topics) == 0 {
		return jsonrpc.NewErrorResponse(c.Request.ID,
			errorsx.New(400, "InvalidParams", "Topics required"))
	}

	result := make(chan []string, 1)
	c.Client.hub.Subscribe <- &SubscribeEvent{
		Client: c.Client,
		Topics: params.Topics,
		Result: result,
	}

	subscribed := <-result
	return jsonrpc.NewResponse(c.Request.ID, map[string]any{
		"subscribed": subscribed,
	})
}

// UnsubscribeHandler handles topic unsubscription.
func UnsubscribeHandler(c *Context) *jsonrpc.Response {
	var params struct {
		Topics []string `json:"topics"`
	}

	if err := c.BindParams(&params); err != nil {
		return jsonrpc.NewErrorResponse(c.Request.ID,
			errorsx.New(400, "InvalidParams", "Invalid unsubscribe params"))
	}

	if len(params.Topics) == 0 {
		return jsonrpc.NewErrorResponse(c.Request.ID,
			errorsx.New(400, "InvalidParams", "Topics required"))
	}

	c.Client.hub.Unsubscribe <- &UnsubscribeEvent{
		Client: c.Client,
		Topics: params.Topics,
	}

	return jsonrpc.NewResponse(c.Request.ID, map[string]any{
		"unsubscribed": params.Topics,
	})
}
