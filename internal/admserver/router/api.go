package router

import (
	"context"

	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/admserver/controller/v1/app"
	"bingo/internal/admserver/controller/v1/syscfg"
	"bingo/internal/admserver/controller/v1/system"
	"bingo/internal/admserver/controller/v1/user"
	"bingo/internal/admserver/middleware"
	"bingo/internal/pkg/store"
	"bingo/pkg/auth"
)

func MapApiRouters(g *gin.Engine) {
	// v1 group
	v1 := g.Group("/v1")

	// Authz
	authz, err := auth.NewAuthz(store.S.DB(context.Background()))
	if err != nil {
		log.Fatalw("auth.NewAuthz error", "err", err)
	}

	authController := system.NewAuthController(store.S, authz)
	adminController := system.NewAdminController(store.S, authz)

	// Login
	v1.POST("auth/login", adminController.Login)

	v1.Use(middleware.Authn())

	// Auth
	v1.GET("auth/user-info", authController.UserInfo)             // 获取登录账号信息
	v1.GET("auth/menus", authController.Menus)                    // 获取登录账号菜单
	v1.PUT("auth/change-password", authController.ChangePassword) // 修改密码
	v1.PUT("auth/switch-role", authController.SwitchRole)         // 切换角色

	v1.Use(middleware.Authz(authz))

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
	v1.POST("menus/:id/toggle-hidden", menuController.ToggleHidden)

	// App Version
	appVersionController := syscfg.NewAppVersionController(store.S, authz)
	v1.GET("cfg/apps", appVersionController.List)
	v1.POST("cfg/apps", appVersionController.Create)
	v1.GET("cfg/apps/:id", appVersionController.Get)
	v1.PUT("cfg/apps/:id", appVersionController.Update)
	v1.DELETE("cfg/apps/:id", appVersionController.Delete)

	// Sys config
	configController := syscfg.NewConfigController(store.S, authz)
	v1.GET("cfg/configs", configController.List)
	v1.POST("cfg/configs", configController.Create)
	v1.GET("cfg/configs/:id", configController.Get)
	v1.PUT("cfg/configs/:id", configController.Update)
	v1.DELETE("cfg/configs/:id", configController.Delete)

	// User
	userController := user.NewUserController(store.S, authz)
	v1.GET("users", userController.List)                                 // 列出用户列表，只有 root 用户才能访问
	v1.POST("users", userController.Create)                              // 创建用户
	v1.GET("users/:name", userController.Get)                            // 获取用户详情
	v1.PUT("users/:name", userController.Update)                         // 更新用户
	v1.DELETE("users/:name", userController.Delete)                      // 删除用户
	v1.PUT("users/:name/change-password", userController.ChangePassword) // 修改用户密码

	// App
	appController := app.NewAppController(store.S, authz)
	v1.GET("apps", appController.List)
	v1.POST("apps", appController.Create)
	v1.GET("apps/:appid", appController.Get)
	v1.PUT("apps/:appid", appController.Update)
	v1.DELETE("apps/:appid", appController.Delete)

	// Api keys
	apiKeyController := app.NewApiKeyController(store.S, authz)
	v1.GET("api-keys", apiKeyController.List)
	v1.POST("api-keys", apiKeyController.Create)
	v1.GET("api-keys/:id", apiKeyController.Get)
	v1.PUT("api-keys/:id", apiKeyController.Update)
	v1.DELETE("api-keys/:id", apiKeyController.Delete)
}
