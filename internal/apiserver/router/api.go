package router

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/http/controller/v1/user"
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
	userController := user.NewUserController(store.S, authz)
	authController := user.NewAuthController(store.S, authz)

	// Login
	v1.POST("login", userController.Login)
	v1.POST("auth/code/email", authController.SendEmailCode)

	// User
	userV1 := v1.Group("users")
	userV1.POST("", userController.Create)                             // 创建用户
	userV1.PUT(":name/change-password", userController.ChangePassword) // 修改用户密码
	userV1.Use(middleware.Authn(), middleware.Authz(authz))
	userV1.GET("", userController.List)           // 列出用户列表，只有 root 用户才能访问
	userV1.GET(":name", userController.Get)       // 获取用户详情
	userV1.PUT(":name", userController.Update)    // 更新用户
	userV1.DELETE(":name", userController.Delete) // 删除用户
}
