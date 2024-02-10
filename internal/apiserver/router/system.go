package router

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/http/controller/v1/system"
	"bingo/internal/apiserver/http/middleware"
	"bingo/internal/apiserver/store"
	"bingo/pkg/auth"
)

func MapSystemRouters(g *gin.Engine) {
	// v1 group
	v1 := g.Group("/v1/system")

	// Authz
	authz, _ := auth.NewAuthz(store.S.DB())
	authController := system.NewAuthController(store.S, authz)
	adminController := system.NewAdminController(store.S, authz)

	// Login
	v1.POST("auth/login", adminController.Login)

	v1.Use(middleware.Authn(), middleware.Authz(authz))

	// Auth
	v1.GET("auth/user-info", authController.UserInfo)             // 获取登录账号信息
	v1.GET("auth/menus", authController.Menus)                    // 获取登录账号菜单
	v1.PUT("auth/change-password", authController.ChangePassword) // 修改密码
	v1.PUT("auth/switch-role", authController.SwitchRole)         // 切换角色

	// Admin
	v1.GET("admins", adminController.List)                                 // 管理员列表
	v1.POST("admins", adminController.Create)                              // 创建管理员
	v1.GET("admins/:name", adminController.Get)                            // 获取管理员详情
	v1.PUT("admins/:name", adminController.Update)                         // 更新管理员信息
	v1.DELETE("admins/:name", adminController.Delete)                      // 删除管理员
	v1.PUT("admins/:name/change-password", adminController.ChangePassword) // 修改密码
	v1.PUT("admins/:name/roles", adminController.SetRoles)                 // 设置角色组

	// Role
	roleController := system.NewRoleController(store.S, authz)
	v1.GET("roles", roleController.List)
	v1.POST("roles", roleController.Create)
	v1.GET("roles/:name", roleController.Get)
	v1.PUT("roles/:name", roleController.Update)
	v1.DELETE("roles/:name", roleController.Delete)
	v1.PUT("roles/:name/apis", roleController.SetApis)     // 设置权限（casbin)
	v1.GET("roles/:name/apis", roleController.GetApiIDs)   // 获取权限 ID 集合（casbin）
	v1.PUT("roles/:name/menus", roleController.SetMenus)   // 设置菜单权限
	v1.GET("roles/:name/menus", roleController.GetMenuIDs) // 获取菜单 ID 集合
	v1.GET("roles/all", roleController.All)

	// API
	apiController := system.NewApiController(store.S, authz)
	v1.GET("apis", apiController.List)
	v1.GET("apis/all", apiController.All)
	v1.POST("apis", apiController.Create)
	v1.GET("apis/:id", apiController.Get)
	v1.PUT("apis/:id", apiController.Update)
	v1.DELETE("apis/:id", apiController.Delete)
	v1.GET("apis/tree", apiController.Tree)

	// Menu
	menuController := system.NewMenuController(store.S, authz)
	v1.GET("menus", menuController.List)
	v1.POST("menus", menuController.Create)
	v1.GET("menus/:id", menuController.Get)
	v1.PUT("menus/:id", menuController.Update)
	v1.DELETE("menus/:id", menuController.Delete)
	v1.GET("menus/tree", menuController.Tree)
}
