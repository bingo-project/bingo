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

		// Admin
		adminM, _ := store.S.Admin().GetUserInfo(c, payload.Subject)
		if adminM.ID == 0 {
			core.WriteResponse(c, errno.ErrTokenInvalid, nil)
			c.Abort()

			return
		}

		var adminInfo v1.AdminInfo
		_ = copier.Copy(&adminInfo, adminM)

		ctx := contextx.WithUserInfo(c.Request.Context(), adminInfo)
		ctx = contextx.WithUserID(ctx, adminInfo.Username)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
