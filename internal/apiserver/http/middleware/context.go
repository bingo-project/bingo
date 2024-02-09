package middleware

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/pkg/auth"
)

// Context is a middleware that injects common prefix fields to gin.Context.
func Context() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(log.KeyTrace, c.GetString(auth.XRequestIDKey))

		c.Next()
	}
}
