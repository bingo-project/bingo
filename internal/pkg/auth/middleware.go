// ABOUTME: HTTP authentication middleware using unified authenticator.
// ABOUTME: Provides Gin middleware for token-based authentication.

package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/pkg/contextx"
	"github.com/bingo-project/bingo/pkg/errorsx"
)

// Middleware returns a Gin middleware that authenticates requests.
func Middleware(a *Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		tokenStr := ExtractBearerToken(authHeader)

		// Verify token and load user
		ctx, err := a.Verify(c.Request.Context(), tokenStr)
		if err != nil {
			e := errorsx.FromError(err)
			core.Response(c, nil, e)
			c.Abort()

			return
		}

		// Update request context and set Gin context values
		c.Request = c.Request.WithContext(ctx)
		c.Set(log.KeySubject, contextx.Username(ctx))
		c.Next()
	}
}
