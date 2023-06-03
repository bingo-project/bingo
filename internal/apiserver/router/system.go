package router

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/controller/v1/system"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/middleware"
	"bingo/pkg/auth"
)

func MapSystemRouters(g *gin.Engine) {
	// v1 group
	v1 := g.Group("/system")

	// Authz
	authz, _ := auth.NewAuthz(store.S.DB())
	adminController := system.NewAdminController(store.S, authz)

	// Login
	v1.POST("login", adminController.Login)

	v1.Use(middleware.Authn(), middleware.Authz(authz))

	// Admin
	v1.GET("admins", adminController.List)            // 管理员列表
	v1.POST("admins", adminController.Create)         // 创建管理员
	v1.GET("admins/:name", adminController.Get)       // 获取管理员详情
	v1.PUT("admins/:name", adminController.Update)    // 更新管理员信息
	v1.DELETE("admins/:name", adminController.Delete) // 删除管理员

	// Role
	roleController := system.NewRoleController(store.S, authz)
	v1.GET("roles", roleController.List)
	v1.POST("roles", roleController.Create)
	v1.GET("roles/:id", roleController.Get)
	v1.PUT("roles/:id", roleController.Update)
	v1.DELETE("roles/:id", roleController.Delete)
}
