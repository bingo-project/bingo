package middleware

import (
	"github.com/bingo-project/component-base/web/token"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/global"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/known"
)

func Authn() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse JWT Token
		payload, err := token.ParseRequest(c)
		if err != nil {
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

			c.Set(known.XUserInfoKey, userInfo)
		}

		// User
		if payload.Info != global.AuthAdmin {
			userInfo, _ := store.S.Users().Get(c, payload.Subject)
			if userInfo.ID == 0 {
				core.WriteResponse(c, errno.ErrTokenInvalid, nil)
				c.Abort()

				return
			}

			c.Set(known.XUserInfoKey, userInfo)
		}

		c.Set(known.XUsernameKey, payload.Subject)
		c.Next()
	}
}
