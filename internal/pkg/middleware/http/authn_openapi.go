package middleware

import (
	"encoding/json"
	"time"

	"github.com/bingo-project/component-base/web/signer"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/known"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/model"
	"bingo/internal/pkg/store"
	"bingo/pkg/util/ip"
)

// AuthnOpenAPI 开放接口校验.
// 1. 时间戳校验
// 2. 随机数校验
// 3. Api key 校验
// - 3.1 是否可用
// - 3.2 是否过期
// - 3.3 IP 是否在白名单中
// 4. 签名校验.
func AuthnOpenAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 是否开启校验
		if !facade.Config.OpenAPI.Enabled {
			c.Next()

			return
		}

		// Get body
		requestBody, _ := c.GetRawData()
		var body map[string]any
		err := json.Unmarshal(requestBody, &body)
		if err != nil {
			core.WriteResponse(c, errno.ErrIllegalRequest, nil)
			c.Abort()

			return
		}

		// Reset request body
		// c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

		// 1. 时间戳校验
		timestamp := body["timestamp"]
		if facade.Config.OpenAPI.TTL > 0 && cast.ToInt64(timestamp) < time.Now().Unix()-facade.Config.OpenAPI.TTL {
			log.C(c).Infow("timestamp error", "timestamp", timestamp, "now", time.Now().Unix())
			core.WriteResponse(c, errno.ErrIllegalRequest, nil)
			c.Abort()

			return
		}

		// 2. 随机数校验
		nonce := cast.ToString(body["nonce"])
		nonceKey := "nonce_" + nonce
		nonceExist := facade.Cache.Has(nonceKey)
		if facade.Config.OpenAPI.Nonce && (nonce == "" || nonceExist) {
			log.C(c).Infow("nonce error", "nonce", nonce)
			core.WriteResponse(c, errno.ErrIllegalRequest, nil)
			c.Abort()

			return
		}

		// 2.1 Cache nonce
		facade.Cache.Set(nonceKey, true, time.Second*time.Duration(facade.Config.OpenAPI.TTL))

		// 3. Api key 校验
		ak := cast.ToString(body["access_key"])
		apiKey, err := store.S.ApiKey().GetByAK(c, ak)
		if err != nil {
			log.C(c).Infow("api key not found", "ak", ak, "err", err)
			core.WriteResponse(c, errno.ErrIllegalRequest, nil)
			c.Abort()

			return
		}

		// 3.1 是否可用
		if apiKey.Status != model.ApiKeyStatusEnabled {
			log.C(c).Infow("api key disabled", "ak", ak)
			core.WriteResponse(c, errno.ErrIllegalRequest, nil)
			c.Abort()

			return
		}

		// 3.2 是否过期
		if apiKey.ExpiredAt != nil && apiKey.ExpiredAt.Before(time.Now()) {
			log.C(c).Infow("api key expired", "ak", ak, "expiredAt", apiKey.ExpiredAt)
			core.WriteResponse(c, errno.ErrIllegalRequest, nil)
			c.Abort()

			return
		}

		// 3.3 IP 是否在白名单中
		forwardedIP := c.GetHeader(known.XForwardedFor)
		clientIP := c.ClientIP()
		if len(apiKey.ACL) > 0 && !ip.ContainsInCIDR(apiKey.ACL, forwardedIP) && !ip.ContainsInCIDR(apiKey.ACL, clientIP) {
			log.C(c).Infow("ip not in whitelist", "ak", ak, "acl", apiKey.ACL, "ip", clientIP, "forwarded", forwardedIP)
			core.WriteResponse(c, errno.ErrIllegalRequest, nil)
			c.Abort()

			return
		}

		// 4. 签名校验
		sign := cast.ToString(body["sign"])
		delete(body, "sign")

		client := signer.New()
		resign := client.Sign(body, apiKey.SecretKey)
		if sign == "" || sign != resign {
			log.C(c).Infow("sign error", "ak", ak, "sign", sign, "resign", resign)
			core.WriteResponse(c, errno.ErrIllegalRequest, nil)
			c.Abort()

			return
		}

		c.Next()
	}
}
