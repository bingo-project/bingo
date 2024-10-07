package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Debug() gin.HandlerFunc {
	return func(c *gin.Context) {
		authStr := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("bingo:bingo")))

		auth := c.Request.Header.Get("Authorization")

		if auth != authStr {
			c.Header("www-Authenticate", "Basic")
			c.AbortWithStatus(http.StatusUnauthorized)

			return
		}

		c.Next()
	}
}
