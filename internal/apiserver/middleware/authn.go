package middleware

import (
	"github.com/bingo-project/component-base/web/token"
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/store"
	"bingo/pkg/contextx"
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

		// User
		userInfo, _ := store.S.User().GetByUID(c, payload.Subject)
		if userInfo.ID == 0 {
			core.WriteResponse(c, errno.ErrTokenInvalid, nil)
			c.Abort()

			return
		}

		ctx := contextx.WithUserInfo(c.Request.Context(), userInfo)
		ctx = contextx.WithUserID(ctx, userInfo.UID)
		ctx = contextx.WithUsername(ctx, userInfo.Username)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
