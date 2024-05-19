package middleware

import (
	"bytes"
	"net/http"
	"time"

	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"bingo/pkg/auth"
)

// 自定义 ResponseWriter 用于捕获响应数据.
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// 重写 Write 方法来捕获数据.
func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建一个新的 responseBodyWriter，并将它设置到上下文的 Writer 中
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		start := time.Now()

		// 调用后续的处理中间件和路由处理函数
		c.Next()

		// Log content
		logFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("referer", c.Request.Referer()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("cost", time.Since(start).Milliseconds()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		}

		// X-Forward
		if forwarded := c.GetHeader(auth.XForwardedKey); forwarded != "" {
			logFields = append(logFields, zap.String(auth.XForwardedKey, forwarded))
		}

		// Log request params & response except GET
		if c.Request.Method != http.MethodGet {
			requestBody, _ := c.GetRawData()
			logFields = append(logFields, zap.String("params", string(requestBody)))
			logFields = append(logFields, zap.String("response", w.body.String()))
		}

		log.C(c).Info("http", logFields...)
	}
}
