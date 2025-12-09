// ABOUTME: Rate limiting middleware for WebSocket handlers.
// ABOUTME: Uses token bucket algorithm with per-method configuration.

package middleware

import (
	"sync"

	"golang.org/x/time/rate"

	"github.com/bingo-project/bingo/pkg/errorsx"
	"github.com/bingo-project/bingo/pkg/jsonrpc"
	"github.com/bingo-project/bingo/pkg/ws"
)

// RateLimitConfig configures rate limiting.
type RateLimitConfig struct {
	Default float64            // Default requests per second (0 = unlimited)
	Methods map[string]float64 // Per-method limits (0 = unlimited)
}

// RateLimiterStore manages per-client rate limiters.
type RateLimiterStore struct {
	mu       sync.RWMutex
	limiters map[*ws.Client]map[string]*rate.Limiter
}

// NewRateLimiterStore creates a new rate limiter store.
func NewRateLimiterStore() *RateLimiterStore {
	return &RateLimiterStore{
		limiters: make(map[*ws.Client]map[string]*rate.Limiter),
	}
}

func (s *RateLimiterStore) get(client *ws.Client, method string, limit float64) *rate.Limiter {
	s.mu.RLock()
	if methods, ok := s.limiters[client]; ok {
		if limiter, ok := methods[method]; ok {
			s.mu.RUnlock()

			return limiter
		}
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.limiters[client] == nil {
		s.limiters[client] = make(map[string]*rate.Limiter)
	}

	if _, ok := s.limiters[client][method]; !ok {
		burst := int(limit) + 1
		if burst < 1 {
			burst = 1
		}
		s.limiters[client][method] = rate.NewLimiter(rate.Limit(limit), burst)
	}

	return s.limiters[client][method]
}

// Remove removes all limiters for a client. Call this on client disconnect.
func (s *RateLimiterStore) Remove(client *ws.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.limiters, client)
}

// RateLimitWithStore creates a rate limit middleware with custom store.
func RateLimitWithStore(cfg *RateLimitConfig, store *RateLimiterStore) ws.Middleware {
	return func(next ws.Handler) ws.Handler {
		return func(c *ws.Context) *jsonrpc.Response {
			if c.Client == nil {
				return next(c)
			}

			// Get limit for this method
			limit := cfg.Default
			if cfg.Methods != nil {
				if methodLimit, ok := cfg.Methods[c.Method]; ok {
					limit = methodLimit
				}
			}

			// No limit
			if limit == 0 {
				return next(c)
			}

			// Check rate limit
			limiter := store.get(c.Client, c.Method, limit)
			if !limiter.Allow() {
				return jsonrpc.NewErrorResponse(c.Request.ID,
					errorsx.New(429, "TooManyRequests", "Rate limit exceeded"))
			}

			return next(c)
		}
	}
}

// RateLimit creates a rate limit middleware with a new store.
// Note: For proper cleanup, use RateLimitWithStore with Hub's WithClientDisconnectCallback.
func RateLimit(cfg *RateLimitConfig) ws.Middleware {
	return RateLimitWithStore(cfg, NewRateLimiterStore())
}

// CleanupClientLimiters is deprecated.
//
// Deprecated: Use RateLimiterStore.Remove with Hub's WithClientDisconnectCallback instead.
func CleanupClientLimiters(client *ws.Client) {
	// no-op for backward compatibility
}
