package middleware

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"net/http"

	"goer/global"
	"goer/global/errno"

	"github.com/gin-gonic/gin"
	"github.com/goer-project/goer/response"
	"github.com/spf13/cast"
	"github.com/ulule/limiter/v3"
	redisStore "github.com/ulule/limiter/v3/drivers/store/redis"
)

func LimitIP(limit string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := resolveRequestSignature(c.ClientIP())
		if ok := handleLimit(c, key, limit); !ok {
			return
		}

		c.Next()
	}
}

// LimitPath Limit ip and url
func LimitPath(limit string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := resolveRequestSignature(c.FullPath() + "|" + c.ClientIP())
		if ok := handleLimit(c, key, limit); !ok {
			return
		}

		c.Next()
	}
}

func LimitMethod(limit string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip GET
		method := c.Request.Method
		if method == http.MethodGet {
			c.Next()
			return
		}

		key := resolveRequestSignature(c.ClientIP())
		if ok := handleLimit(c, key, limit); !ok {
			return
		}

		c.Next()
	}
}

func resolveRequestSignature(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	bs := h.Sum(nil)

	return hex.EncodeToString(bs)
}

func handleLimit(c *gin.Context, key string, limit string) bool {
	rate, err := GetLimiterContext(c, key, limit)
	if err != nil {
		log.Println(err)
		response.Fail(c, errno.InternalServerError)
		return false
	}

	// Response Headers
	c.Header("X-RateLimit-Limit", cast.ToString(rate.Limit))         // 单位时间的访问上限
	c.Header("X-RateLimit-Remaining", cast.ToString(rate.Remaining)) // 剩余的访问次数
	c.Header("X-RateLimit-Reset", cast.ToString(rate.Reset))         // 访问次数重置时间

	if !rate.Reached {
		return true
	}

	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"message": "Too Many Requests",
	})

	return false
}

func GetLimiterContext(c *gin.Context, key string, formatted string) (limiter.Context, error) {
	var context limiter.Context

	// Define a limit rate to 4 requests per hour.
	rate, err := limiter.NewRateFromFormatted(formatted)
	if err != nil {
		return context, err
	}

	// Create a store with the redis client.
	store, err := redisStore.NewStoreWithOptions(global.Redis.Client, limiter.StoreOptions{
		Prefix: global.Config.App.Name + ":limiter",
	})
	if err != nil {
		return context, err
	}

	// New limiter instance
	instance := limiter.New(store, rate)

	return instance.Get(c, key)
}
