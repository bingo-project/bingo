package router

import (
	"context"

	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	auth2 "bingo/internal/apiserver/handler/http/auth"
	"bingo/internal/apiserver/middleware"
	"bingo/internal/pkg/store"
	"bingo/pkg/auth"
)

func MapApiRouters(g *gin.Engine) {
	// v1 group
	v1 := g.Group("/v1")
	v1.Use(middleware.Maintenance())

	// Authz
	authz, err := auth.NewAuthz(store.S.DB(context.Background()))
	if err != nil {
		log.Fatalw("auth.NewAuthz error", "err", err)
	}

	authController := auth2.NewAuthController(store.S, authz)

	// Login
	v1.POST("auth/code/email", authController.SendEmailCode)
	v1.POST("auth/register", authController.Register)
	v1.POST("auth/login", authController.Login)

	// Login by Address
	v1.GET("auth/nonce", authController.Nonce)
	v1.POST("auth/login/address", authController.LoginByAddress)

	// Login by Third Party
	v1.GET("auth/providers", authController.Providers)
	v1.GET("auth/login/:provider", authController.GetAuthCode)
	v1.POST("auth/login/:provider", authController.LoginByProvider)

	v1.Use(middleware.Authn())

	// Auth
	v1.GET("auth/user-info", authController.UserInfo)             // 获取登录账号信息
	v1.PUT("auth/change-password", authController.ChangePassword) // 修改用户密码
	v1.POST("auth/bind/:provider", authController.BindProvider)
}
