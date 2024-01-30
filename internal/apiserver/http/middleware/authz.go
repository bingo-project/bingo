package middleware

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/global"
	"bingo/internal/apiserver/model"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/pkg/auth"
)

// Author 用来定义授权接口实现.
// sub: 操作主题，obj：操作对象, act：操作
type Author interface {
	Authorize(sub, obj, act string) (bool, error)
}

// Authz 是 Gin 中间件，用来进行请求授权.
func Authz(a Author) gin.HandlerFunc {
	return func(c *gin.Context) {
		// User
		sub := c.GetString(auth.XUsernameKey)
		obj := c.Request.URL.Path
		act := c.Request.Method

		// System admin
		guard := c.GetString(auth.XGuard)
		if guard == global.AuthAdmin {
			var admin model.AdminM
			err := auth.User(c, &admin)
			if err != nil {
				core.WriteResponse(c, errno.ErrUnauthorized, nil)

				return
			}

			sub = global.RolePrefix + admin.RoleName
		}

		log.C(c).Debugw("Build authorize context", "sub", sub, "obj", obj, "act", act)
		if allowed, _ := a.Authorize(sub, obj, act); !allowed {
			core.WriteResponse(c, errno.ErrForbidden, nil)
			c.Abort()

			return
		}
	}
}
