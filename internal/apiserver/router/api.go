package router

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/global"
	auth2 "bingo/internal/apiserver/http/controller/v1/auth"
	"bingo/internal/apiserver/http/middleware"
	"bingo/internal/apiserver/store"
	"bingo/pkg/auth"
)

func MapApiRouters(g *gin.Engine) {
	// v1 group
	v1 := g.Group("/v1")
	v1.Use(middleware.Maintenance())

	// Authz
	authz, _ := auth.NewAuthz(store.S.DB())
	authController := auth2.NewAuthController(store.S, authz)

	// Login
	v1.POST("auth/code/email", authController.SendEmailCode)
	v1.POST("auth/register", authController.Register)
	v1.POST("auth/login", authController.Login)

	v1.Use(middleware.Authn(global.AuthUser))

	// Auth
	v1.GET("auth/user-info", authController.UserInfo)             // 获取登录账号信息
	v1.PUT("auth/change-password", authController.ChangePassword) // 修改用户密码
}
