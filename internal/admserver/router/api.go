package router

import (
	"context"

	"github.com/gin-gonic/gin"

	bizauth "bingo/internal/admserver/biz/auth"
	"bingo/internal/admserver/handler/http/app"
	"bingo/internal/admserver/handler/http/config"
	"bingo/internal/admserver/handler/http/system"
	"bingo/internal/admserver/handler/http/user"
	"bingo/internal/pkg/auth"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/store"
	pkgauth "bingo/pkg/auth"
)

func MapApiRouters(g *gin.Engine) {
	// v1 group
	v1 := g.Group("/v1")

	// Authz (still using pkg/auth for Casbin policy management)
	authz, err := pkgauth.NewAuthz(store.S.DB(context.Background()))
	if err != nil {
		log.Fatalw("auth.NewAuthz error", "err", err)
	}

	authHandler := system.NewAuthHandler(store.S, authz)
	adminHandler := system.NewAdminHandler(store.S, authz)

	// Login
	v1.POST("auth/login", adminHandler.Login)

	// Authentication middleware
	loader := bizauth.NewAdminLoader(store.S)
	authn := auth.New(loader)
	v1.Use(auth.Middleware(authn))

	// Auth
	v1.GET("auth/user-info", authHandler.UserInfo)             // 获取登录账号信息
	v1.GET("auth/menus", authHandler.Menus)                    // 获取登录账号菜单
	v1.PUT("auth/change-password", authHandler.ChangePassword) // 修改密码
	v1.PUT("auth/switch-role", authHandler.SwitchRole)         // 切换角色

	// Authorization middleware
	resolver := &bizauth.AdminSubjectResolver{}
	authorizer, err := auth.NewAuthorizer(store.S.DB(context.Background()), resolver)
	if err != nil {
		log.Fatalw("auth.NewAuthorizer error", "err", err)
	}
	v1.Use(auth.AuthzMiddleware(authorizer))

	// Admin
	v1.GET("admins", adminHandler.List)                                 // 管理员列表
	v1.POST("admins", adminHandler.Create)                              // 创建管理员
	v1.GET("admins/:name", adminHandler.Get)                            // 获取管理员详情
	v1.PUT("admins/:name", adminHandler.Update)                         // 更新管理员信息
	v1.DELETE("admins/:name", adminHandler.Delete)                      // 删除管理员
	v1.PUT("admins/:name/change-password", adminHandler.ChangePassword) // 修改密码
	v1.PUT("admins/:name/roles", adminHandler.SetRoles)                 // 设置角色组

	// Role
	roleHandler := system.NewRoleHandler(store.S, authz)
	v1.GET("roles", roleHandler.List)
	v1.POST("roles", roleHandler.Create)
	v1.GET("roles/:name", roleHandler.Get)
	v1.PUT("roles/:name", roleHandler.Update)
	v1.DELETE("roles/:name", roleHandler.Delete)
	v1.PUT("roles/:name/apis", roleHandler.SetApis)     // 设置权限（casbin)
	v1.GET("roles/:name/apis", roleHandler.GetApiIDs)   // 获取权限 ID 集合（casbin）
	v1.PUT("roles/:name/menus", roleHandler.SetMenus)   // 设置菜单权限
	v1.GET("roles/:name/menus", roleHandler.GetMenuIDs) // 获取菜单 ID 集合
	v1.GET("roles/all", roleHandler.All)

	// API
	apiHandler := system.NewApiHandler(store.S, authz)
	v1.GET("apis", apiHandler.List)
	v1.GET("apis/all", apiHandler.All)
	v1.POST("apis", apiHandler.Create)
	v1.GET("apis/:id", apiHandler.Get)
	v1.PUT("apis/:id", apiHandler.Update)
	v1.DELETE("apis/:id", apiHandler.Delete)
	v1.GET("apis/tree", apiHandler.Tree)

	// Menu
	menuHandler := system.NewMenuHandler(store.S, authz)
	v1.GET("menus", menuHandler.List)
	v1.POST("menus", menuHandler.Create)
	v1.GET("menus/:id", menuHandler.Get)
	v1.PUT("menus/:id", menuHandler.Update)
	v1.DELETE("menus/:id", menuHandler.Delete)
	v1.GET("menus/tree", menuHandler.Tree)
	v1.POST("menus/:id/toggle-hidden", menuHandler.ToggleHidden)

	// App Version
	appVersionHandler := config.NewAppVersionHandler(store.S, authz)
	v1.GET("cfg/apps", appVersionHandler.List)
	v1.POST("cfg/apps", appVersionHandler.Create)
	v1.GET("cfg/apps/:id", appVersionHandler.Get)
	v1.PUT("cfg/apps/:id", appVersionHandler.Update)
	v1.DELETE("cfg/apps/:id", appVersionHandler.Delete)

	// Sys config
	configHandler := config.NewConfigHandler(store.S, authz)
	v1.GET("cfg/configs", configHandler.List)
	v1.POST("cfg/configs", configHandler.Create)
	v1.GET("cfg/configs/:id", configHandler.Get)
	v1.PUT("cfg/configs/:id", configHandler.Update)
	v1.DELETE("cfg/configs/:id", configHandler.Delete)

	// User
	userHandler := user.NewUserHandler(store.S, authz)
	v1.GET("users", userHandler.List)                                 // 列出用户列表，只有 root 用户才能访问
	v1.POST("users", userHandler.Create)                              // 创建用户
	v1.GET("users/:name", userHandler.Get)                            // 获取用户详情
	v1.PUT("users/:name", userHandler.Update)                         // 更新用户
	v1.DELETE("users/:name", userHandler.Delete)                      // 删除用户
	v1.PUT("users/:name/change-password", userHandler.ChangePassword) // 修改用户密码

	// App
	appHandler := app.NewAppHandler(store.S, authz)
	v1.GET("apps", appHandler.List)
	v1.POST("apps", appHandler.Create)
	v1.GET("apps/:appid", appHandler.Get)
	v1.PUT("apps/:appid", appHandler.Update)
	v1.DELETE("apps/:appid", appHandler.Delete)

	// Api keys
	apiKeyHandler := app.NewApiKeyHandler(store.S, authz)
	v1.GET("api-keys", apiKeyHandler.List)
	v1.POST("api-keys", apiKeyHandler.Create)
	v1.GET("api-keys/:id", apiKeyHandler.Get)
	v1.PUT("api-keys/:id", apiKeyHandler.Update)
	v1.DELETE("api-keys/:id", apiKeyHandler.Delete)
}
