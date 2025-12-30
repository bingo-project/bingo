// ABOUTME: AI rate limiter middleware.
// ABOUTME: Limits AI requests per user based on RPM quota using Redis.

package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/pkg/contextx"
)

// AILimiter creates a rate limiter middleware for AI endpoints.
// Uses Redis for distributed rate limiting.
func AILimiter(defaultRPM int) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := contextx.UserID(c)
		if uid == "" {
			c.Next()

			return
		}

		// Get user's RPM limit (default for now, can be from DB)
		rpm := defaultRPM
		if rpm <= 0 {
			rpm = 10 // fallback default
		}

		// Use Redis-based limiter via shared GetLimiterContext
		key := fmt.Sprintf("ai:rpm:%s", uid)
		limit := fmt.Sprintf("%d-M", rpm) // e.g., "20-M" = 20 per minute

		ctx, err := GetLimiterContext(c, key, limit)
		if err != nil {
			core.Response(c, nil, errno.ErrOperationFailed.WithMessage("rate limiter error: %v", err))
			c.Abort()

			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", ctx.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", ctx.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", ctx.Reset))

		if ctx.Reached {
			core.Response(c, nil, errno.ErrAIQuotaExceeded)
			c.Abort()

			return
		}

		c.Next()
	}
}
