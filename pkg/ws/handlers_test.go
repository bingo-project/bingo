// ABOUTME: Tests for built-in WebSocket handlers.
// ABOUTME: Validates heartbeat, subscribe, and unsubscribe functionality.

package ws

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bingo-project/bingo/pkg/jsonrpc"
)

func TestHeartbeatHandler(t *testing.T) {
	c := &Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "heartbeat"},
		Client:  &Client{HeartbeatTime: 0},
		Method:  "heartbeat",
	}

	before := time.Now().Unix()
	resp := HeartbeatHandler(c)
	after := time.Now().Unix()

	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]any)
	assert.Equal(t, "ok", result["status"])

	serverTime := result["server_time"].(int64)
	assert.GreaterOrEqual(t, serverTime, before)
	assert.LessOrEqual(t, serverTime, after)
}

func TestSubscribeHandler(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	client := &Client{
		ID:        "test",
		UserID:    "user-1",
		Platform:  "web",
		LoginTime: 1000,
		hub:       hub,
		Send:      make(chan []byte, 256),
	}

	params, _ := json.Marshal(map[string][]string{"topics": {"market.BTC"}})
	c := &Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "subscribe", Params: params},
		Client:  client,
		Method:  "subscribe",
	}

	resp := SubscribeHandler(c)

	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]any)
	subscribed := result["subscribed"].([]string)
	assert.Contains(t, subscribed, "market.BTC")
}

func TestSubscribeHandler_InvalidParams(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:  "test",
		hub: hub,
	}

	c := &Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "subscribe", Params: json.RawMessage(`invalid`)},
		Client:  client,
		Method:  "subscribe",
	}

	resp := SubscribeHandler(c)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, -32602, resp.Error.Code) // JSON-RPC Invalid Params
}

func TestSubscribeHandler_EmptyTopics(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:  "test",
		hub: hub,
	}

	params, _ := json.Marshal(map[string][]string{"topics": {}})
	c := &Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "subscribe", Params: params},
		Client:  client,
		Method:  "subscribe",
	}

	resp := SubscribeHandler(c)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, -32602, resp.Error.Code) // JSON-RPC Invalid Params
}

func TestUnsubscribeHandler(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	client := &Client{
		ID:        "test",
		UserID:    "user-1",
		Platform:  "web",
		LoginTime: 1000,
		hub:       hub,
		Send:      make(chan []byte, 256),
	}

	params, _ := json.Marshal(map[string][]string{"topics": {"market.BTC"}})
	c := &Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "unsubscribe", Params: params},
		Client:  client,
		Method:  "unsubscribe",
	}

	resp := UnsubscribeHandler(c)

	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]any)
	unsubscribed := result["unsubscribed"].([]string)
	assert.Contains(t, unsubscribed, "market.BTC")
}

func TestUnsubscribeHandler_InvalidParams(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:  "test",
		hub: hub,
	}

	c := &Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "unsubscribe", Params: json.RawMessage(`invalid`)},
		Client:  client,
		Method:  "unsubscribe",
	}

	resp := UnsubscribeHandler(c)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, -32602, resp.Error.Code) // JSON-RPC Invalid Params
}

func TestUnsubscribeHandler_EmptyTopics(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:  "test",
		hub: hub,
	}

	params, _ := json.Marshal(map[string][]string{"topics": {}})
	c := &Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "unsubscribe", Params: params},
		Client:  client,
		Method:  "unsubscribe",
	}

	resp := UnsubscribeHandler(c)

	assert.NotNil(t, resp.Error)
	assert.Equal(t, -32602, resp.Error.Code) // JSON-RPC Invalid Params
}
