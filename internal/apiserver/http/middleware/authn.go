package middleware

import (
	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/web/token"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/apiserver/global"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/pkg/auth"
)

func Authn(guard ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse JWT Token
		payload, err := token.ParseRequest(c.Request)
		if err != nil {
			core.WriteResponse(c, errno.ErrTokenInvalid, nil)
			c.Abort()

			return
		}

		// Check guard
		if len(guard) > 0 && guard[0] != "" && cast.ToString(payload.Info) != guard[0] {
			core.WriteResponse(c, errno.ErrTokenInvalid, nil)
			c.Abort()

			return
		}

		// Admin
		if payload.Info == global.AuthAdmin {
			userInfo, _ := store.S.Admins().GetUserInfo(c, payload.Subject)
			if userInfo.ID == 0 {
				core.WriteResponse(c, errno.ErrTokenInvalid, nil)
				c.Abort()

				return
			}

			c.Set(auth.XUserInfoKey, userInfo)
		}

		// User
		if payload.Info != global.AuthAdmin {
			userInfo, _ := store.S.Users().GetByUID(c, payload.Subject)
			if userInfo.ID == 0 {
				core.WriteResponse(c, errno.ErrTokenInvalid, nil)
				c.Abort()

				return
			}

			c.Set(auth.XUserInfoKey, userInfo)
		}

		c.Set(auth.XGuard, payload.Info)
		c.Set(auth.XUsernameKey, payload.Subject)
		c.Set(log.KeySubject, payload.Subject)
		c.Next()
	}
}
