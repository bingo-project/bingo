// ABOUTME: Rate limiting middleware for WebSocket handlers.
// ABOUTME: Uses token bucket algorithm with per-method configuration.

package middleware

import (
	"sync"

	"golang.org/x/time/rate"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
	"bingo/pkg/ws"
)

// RateLimitConfig configures rate limiting.
type RateLimitConfig struct {
	Default float64            // Default requests per second (0 = unlimited)
	Methods map[string]float64 // Per-method limits (0 = unlimited)
}

// clientLimiters stores per-client limiters.
type clientLimiters struct {
	mu       sync.RWMutex
	limiters map[*ws.Client]map[string]*rate.Limiter
}

var limiters = &clientLimiters{
	limiters: make(map[*ws.Client]map[string]*rate.Limiter),
}

func (cl *clientLimiters) get(client *ws.Client, method string, limit float64) *rate.Limiter {
	cl.mu.RLock()
	if methods, ok := cl.limiters[client]; ok {
		if limiter, ok := methods[method]; ok {
			cl.mu.RUnlock()
			return limiter
		}
	}
	cl.mu.RUnlock()

	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.limiters[client] == nil {
		cl.limiters[client] = make(map[string]*rate.Limiter)
	}

	if _, ok := cl.limiters[client][method]; !ok {
		burst := int(limit) + 1
		if burst < 1 {
			burst = 1
		}
		cl.limiters[client][method] = rate.NewLimiter(rate.Limit(limit), burst)
	}

	return cl.limiters[client][method]
}

func (cl *clientLimiters) remove(client *ws.Client) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	delete(cl.limiters, client)
}

// RateLimit limits request rate per client per method.
func RateLimit(cfg *RateLimitConfig) ws.Middleware {
	return func(next ws.Handler) ws.Handler {
		return func(mc *ws.MiddlewareContext) *jsonrpc.Response {
			if mc.Client == nil {
				return next(mc)
			}

			// Get limit for this method
			limit := cfg.Default
			if cfg.Methods != nil {
				if methodLimit, ok := cfg.Methods[mc.Method]; ok {
					limit = methodLimit
				}
			}

			// No limit
			if limit == 0 {
				return next(mc)
			}

			// Check rate limit
			limiter := limiters.get(mc.Client, mc.Method, limit)
			if !limiter.Allow() {
				return jsonrpc.NewErrorResponse(mc.Request.ID,
					errorsx.New(429, "TooManyRequests", "Rate limit exceeded"))
			}

			return next(mc)
		}
	}
}

// CleanupClientLimiters removes limiters for a disconnected client.
// Call this when client disconnects.
func CleanupClientLimiters(client *ws.Client) {
	limiters.remove(client)
}
