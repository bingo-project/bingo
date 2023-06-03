package middleware

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/known"
)

// UsernameKey defines the key in gin context which represents the owner of the secret.
const (
	KeyRequestID string = "requestID"
	KeyUsername  string = "username"
)

// Context is a middleware that injects common prefix fields to gin.Context.
func Context() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(KeyRequestID, c.GetString(known.XRequestIDKey))
		c.Set(KeyUsername, c.GetString(KeyUsername))
		c.Next()
	}
}
