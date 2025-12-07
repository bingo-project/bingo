// ABOUTME: HTTP authentication middleware using unified authenticator.
// ABOUTME: Provides Gin middleware for token-based authentication.

package auth

import (
	"github.com/bingo-project/component-base/web/token"
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/log"
	"bingo/pkg/contextx"
	"bingo/pkg/errorsx"
)

// Middleware returns a Gin middleware that authenticates requests.
func Middleware(a *Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		tokenStr := ExtractBearerToken(authHeader)

		// Verify token
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

// MiddlewareFromRequest returns a Gin middleware that uses token.ParseRequest.
// This maintains compatibility with the existing token parsing approach.
func MiddlewareFromRequest(a *Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse JWT Token using component-base parser
		payload, err := token.ParseRequest(c.Request)
		if err != nil {
			core.Response(c, nil, errorsx.New(401, "Unauthenticated", "invalid token: %s", err.Error()))
			c.Abort()

			return
		}

		// Set context values
		ctx := c.Request.Context()
		ctx = contextx.WithUserID(ctx, payload.Subject)
		ctx = contextx.WithUsername(ctx, payload.Subject)

		c.Request = c.Request.WithContext(ctx)
		c.Set(log.KeySubject, payload.Subject)
		c.Next()
	}
}
