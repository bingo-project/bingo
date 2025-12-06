package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// NoCache is a middleware function that appends headers
// to prevent the client from caching the HTTP response.
func NoCache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate, value")
	c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	c.Next()
}

// Cors adds CORS headers for Gin handlers.
// Note: CorsHandler provides the same functionality for net/http handlers.
func Cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	c.Header("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")

	if c.Request.Method == "OPTIONS" {
		c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.AbortWithStatus(http.StatusOK)

		return
	}

	c.Next()
}

// CorsHandler wraps an http.Handler with CORS headers.
// Note: Cors provides the same functionality for Gin handlers.
func CorsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)

			return
		}

		h.ServeHTTP(w, r)
	})
}

// Secure is a middleware function that appends security
// and resource access headers.
func Secure(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-XSS-Protection", "1; mode=block")

	if c.Request.TLS != nil {
		c.Header("Strict-Transport-Security", "max-age=31536000")
	}
}

func ForceUserAgent(c *gin.Context) {
	if len(c.Request.Header["User-Agent"]) > 0 {
		c.Next()

		return
	}

	_ = c.AbortWithError(http.StatusBadRequest, errors.New("User-Agent not found"))
}
