package middleware

import (
	"github.com/bingo-project/component-base/web/token"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1"
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
		userM, _ := store.S.User().GetByUID(c, payload.Subject)
		if userM.ID == 0 {
			core.WriteResponse(c, errno.ErrTokenInvalid, nil)
			c.Abort()

			return
		}

		var userInfo v1.UserInfo
		_ = copier.Copy(&userInfo, userM)
		userInfo.PayPassword = userM.PayPassword != ""

		ctx := contextx.WithUserInfo(c.Request.Context(), &userInfo)
		ctx = contextx.WithUserID(ctx, userInfo.UID)
		ctx = contextx.WithUsername(ctx, userInfo.Username)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
