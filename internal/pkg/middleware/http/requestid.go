package middleware

import (
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"bingo/internal/pkg/known"
	"bingo/pkg/contextx"
)

// RequestID is a middleware that injects a 'X-Request-ID' into the context and request/response header of each request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for incoming header, use it if exists
		rid := c.GetHeader(known.XRequestID)
		if rid == "" {
			rid = uuid.New().String()
		}

		// Set request id to context & response header.
		ctx := contextx.WithRequestID(c.Request.Context(), rid)
		c.Request = c.Request.WithContext(ctx)

		c.Writer.Header().Set(known.XRequestID, rid)

		c.Next()
	}
}

// ClientIP is a middleware that injects common prefix fields to context.
func ClientIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = contextx.WithClientIP(ctx, c.ClientIP())
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetLoggerConfig return gin.LoggerConfig which will write the logs to specified io.Writer with given gin.LogFormatter.
// By default gin.DefaultWriter = os.Stdout
// reference: https://github.com/gin-gonic/gin#custom-log-format
func GetLoggerConfig(formatter gin.LogFormatter, output io.Writer, skipPaths []string) gin.LoggerConfig {
	if formatter == nil {
		formatter = GetDefaultLogFormatterWithRequestID()
	}

	return gin.LoggerConfig{
		Formatter: formatter,
		Output:    output,
		SkipPaths: skipPaths,
	}
}

// GetDefaultLogFormatterWithRequestID returns gin.LogFormatter with 'RequestID'.
func GetDefaultLogFormatterWithRequestID() gin.LogFormatter {
	return func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			// Truncate in a golang < 1.8 safe way
			param.Latency -= param.Latency % time.Second
		}

		return fmt.Sprintf("%s%3d%s - [%s] \"%v %s%s%s %s\" %s",
			// param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			statusColor, param.StatusCode, resetColor,
			param.ClientIP,
			param.Latency,
			methodColor, param.Method, resetColor,
			param.Path,
			param.ErrorMessage,
		)
	}
}

// GetRequestIDFromContext returns 'RequestID' from the given context if present.
func GetRequestIDFromContext(c *gin.Context) string {
	return contextx.RequestID(c.Request.Context())
}

// GetRequestIDFromHeaders returns 'RequestID' from the headers if present.
func GetRequestIDFromHeaders(c *gin.Context) string {
	return c.Request.Header.Get(known.XRequestID)
}
