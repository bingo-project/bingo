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
	v1 := g.Group("/v1/system")

	// Authz
	authz, _ := auth.NewAuthz(store.S.DB())
	adminController := system.NewAdminController(store.S, authz)

	// Login
	v1.POST("login", adminController.Login)

	v1.Use(middleware.Authn(), middleware.Authz(authz))

	// Admin
	v1.GET("admins", adminController.List)                                 // 管理员列表
	v1.POST("admins", adminController.Create)                              // 创建管理员
	v1.GET("admins/:name", adminController.Get)                            // 获取管理员详情
	v1.GET("admins/self", adminController.Self)                            // 获取登录账号信息
	v1.PUT("admins/:name", adminController.Update)                         // 更新管理员信息
	v1.DELETE("admins/:name", adminController.Delete)                      // 删除管理员
	v1.PUT("admins/:name/change-password", adminController.ChangePassword) // 修改密码
	v1.PUT("admins/:name/roles", adminController.SetRoles)                 // 设置角色组
	v1.PUT("admins/:name/switch-role", adminController.SwitchRole)         // 切换角色

	// Role
	roleController := system.NewRoleController(store.S, authz)
	v1.GET("roles", roleController.List)
	v1.POST("roles", roleController.Create)
	v1.GET("roles/:name", roleController.Get)
	v1.PUT("roles/:name", roleController.Update)
	v1.DELETE("roles/:name", roleController.Delete)
	v1.PUT("roles/:name/apis", roleController.SetApis)   // 设置权限（casbin)
	v1.GET("roles/:name/apis", roleController.GetApiIDs) // 获取权限ID（casbin）

	// API
	apiController := system.NewApiController(store.S, authz)
	v1.GET("apis", apiController.List)
	v1.GET("apis/all", apiController.All)
	v1.POST("apis", apiController.Create)
	v1.GET("apis/:id", apiController.Get)
	v1.PUT("apis/:id", apiController.Update)
	v1.DELETE("apis/:id", apiController.Delete)
}
