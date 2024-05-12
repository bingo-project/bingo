package router

import (
	"github.com/gin-gonic/gin"

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

	v1.Use(middleware.Authn(), middleware.Authz(authz))

	// Auth
	v1.PUT(":name/change-password", authController.ChangePassword) // 修改用户密码
}
