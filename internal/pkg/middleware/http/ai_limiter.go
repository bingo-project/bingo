// ABOUTME: AI rate limiter middleware.
// ABOUTME: Limits AI requests per user based on RPM quota.

package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/pkg/contextx"
)

// AILimiter creates a rate limiter middleware for AI endpoints
func AILimiter(defaultRPM int) gin.HandlerFunc {
	// Use in-memory store for now; can switch to Redis for distributed
	store := memory.NewStore()

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

		// Create rate with limit
		rate := limiter.Rate{
			Period: time.Minute,
			Limit:  int64(rpm),
		}

		limiterInstance := limiter.New(store, rate)
		key := fmt.Sprintf("ai:%s", uid)

		context, err := limiterInstance.Get(c, key)
		if err != nil {
			core.Response(c, nil, errno.ErrOperationFailed.WithMessage("rate limiter error: %v", err))
			c.Abort()

			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

		if context.Reached {
			core.Response(c, nil, errno.ErrAIQuotaExceeded)
			c.Abort()

			return
		}

		c.Next()
	}
}
