package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ErrLimitExceeded defines Limit exceeded error.
var ErrLimitExceeded = errors.New("Limit exceeded")

// Limit drops (HTTP status 429) the request if the limit is reached.
func Limit(maxEventsPerSec float64, maxBurstSize int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(maxEventsPerSec), maxBurstSize)

	return func(c *gin.Context) {
		if limiter.Allow() {
			c.Next()

			return
		}

		// Limit reached
		_ = c.Error(ErrLimitExceeded)
		c.AbortWithStatus(429)
	}
}
