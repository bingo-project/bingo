// ABOUTME: Tests for rate limiting middleware.
// ABOUTME: Verifies rate limiting with token bucket algorithm.

package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

func TestRateLimit_Allows(t *testing.T) {
	cfg := &RateLimitConfig{
		Default: 10, // 10 requests per second
	}

	handler := func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := RateLimit(cfg)(handler)

	client := &ws.Client{
		Addr: "127.0.0.1:12345",
	}
	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  client,
		Method:  "test",
	}

	resp := wrapped(c)

	assert.Nil(t, resp.Error)
	assert.Equal(t, "ok", resp.Result)

	// Cleanup
	CleanupClientLimiters(client)
}

func TestRateLimit_Blocks(t *testing.T) {
	cfg := &RateLimitConfig{
		Default: 1, // 1 request per second
	}

	handler := func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := RateLimit(cfg)(handler)

	client := &ws.Client{
		Addr: "127.0.0.1:12346",
	}
	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  client,
		Method:  "test",
	}

	// First request should succeed (uses burst token)
	resp := wrapped(c)
	assert.Nil(t, resp.Error)

	// Second request also succeeds (burst = limit + 1 = 2)
	resp = wrapped(c)
	assert.Nil(t, resp.Error)

	// Third request should fail (burst exhausted)
	resp = wrapped(c)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "TooManyRequests", resp.Error.Reason)

	// Cleanup
	CleanupClientLimiters(client)
}

func TestRateLimit_MethodSpecific(t *testing.T) {
	cfg := &RateLimitConfig{
		Default: 1,
		Methods: map[string]float64{
			"heartbeat": 0, // No limit
		},
	}

	handler := func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := RateLimit(cfg)(handler)

	client := &ws.Client{
		Addr: "127.0.0.1:12347",
	}

	// Heartbeat should always succeed
	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "heartbeat"},
		Client:  client,
		Method:  "heartbeat",
	}

	for i := 0; i < 10; i++ {
		resp := wrapped(c)
		assert.Nil(t, resp.Error, "heartbeat %d should succeed", i)
	}

	// Cleanup
	CleanupClientLimiters(client)
}

func TestRateLimit_NilClient(t *testing.T) {
	cfg := &RateLimitConfig{
		Default: 1,
	}

	handler := func(c *ws.Context) *jsonrpc.Response {
		return jsonrpc.NewResponse(c.Request.ID, "ok")
	}

	wrapped := RateLimit(cfg)(handler)

	c := &ws.Context{
		Context: context.Background(),
		Request: &jsonrpc.Request{ID: 1, Method: "test"},
		Client:  nil,
		Method:  "test",
	}

	// Should pass through without rate limiting
	resp := wrapped(c)
	assert.Nil(t, resp.Error)
}
