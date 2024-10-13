package middleware

import (
	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/web/token"
	"github.com/gin-gonic/gin"

	"bingo/internal/admserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/pkg/auth"
)

func Authn() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse JWT Token
		payload, err := token.ParseRequest(c.Request)
		if err != nil {
			core.WriteResponse(c, errno.ErrTokenInvalid, nil)
			c.Abort()

			return
		}

		// Admin
		userInfo, _ := store.S.Admins().GetUserInfo(c, payload.Subject)
		if userInfo.ID == 0 {
			core.WriteResponse(c, errno.ErrTokenInvalid, nil)
			c.Abort()

			return
		}

		c.Set(auth.XUserInfoKey, userInfo)
		c.Set(auth.XUsernameKey, payload.Subject)
		c.Set(log.KeySubject, payload.Subject)
		c.Next()
	}
}
