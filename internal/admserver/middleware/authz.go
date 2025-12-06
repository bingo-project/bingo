package middleware

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/global"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/model"
	"bingo/pkg/contextx"
)

// Author 用来定义授权接口实现.
// sub: 操作主题，obj：操作对象, act：操作
type Author interface {
	Authorize(sub, obj, act string) (bool, error)
}

// Authz 是 Gin 中间件，用来进行请求授权.
func Authz(a Author) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// User
		sub := contextx.Username(ctx)
		obj := c.Request.URL.Path
		act := c.Request.Method

		// System admin
		admin, ok := contextx.UserInfo[*model.AdminM](ctx)
		if !ok {
			core.WriteResponse(c, errno.ErrUnauthorized, nil)

			return
		}

		sub = global.RolePrefix + admin.RoleName

		log.C(c).Debugw("Build authorize context", "sub", sub, "obj", obj, "act", act)
		if allowed, _ := a.Authorize(sub, obj, act); !allowed {
			core.WriteResponse(c, errno.ErrForbidden, nil)
			c.Abort()

			return
		}
	}
}
