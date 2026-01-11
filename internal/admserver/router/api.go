package router

import (
	"context"

	"github.com/gin-gonic/gin"

	bizauth "github.com/bingo-project/bingo/internal/admserver/biz/auth"
	"github.com/bingo-project/bingo/internal/admserver/handler/http/ai"
	"github.com/bingo-project/bingo/internal/admserver/handler/http/app"
	handlerauth "github.com/bingo-project/bingo/internal/admserver/handler/http/auth"
	"github.com/bingo-project/bingo/internal/admserver/handler/http/config"
	"github.com/bingo-project/bingo/internal/admserver/handler/http/notification"
	"github.com/bingo-project/bingo/internal/admserver/handler/http/system"
	"github.com/bingo-project/bingo/internal/admserver/handler/http/user"
	aipkg "github.com/bingo-project/bingo/internal/pkg/ai"
	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

func MapApiRouters(g *gin.Engine) {
	// v1 group
	v1 := g.Group("/v1")

	// Authorizer for policy management (resolver is nil since these handlers only manage policies)
	policyAuthz, err := auth.NewAuthorizer(store.S.DB(context.Background()), nil)
	if err != nil {
		log.Fatalw("auth.NewAuthorizer error", "err", err)
	}

	authHandler := system.NewAuthHandler(store.S, policyAuthz)
	adminHandler := system.NewAdminHandler(store.S, policyAuthz)

	// Login
	v1.POST("auth/login", adminHandler.Login)
	v1.POST("auth/login/totp", adminHandler.LoginWithTOTP)

	// Authentication middleware
	loader := bizauth.NewAdminLoader(store.S)
	authn := auth.New(loader)
	v1.Use(auth.Middleware(authn))

	// Auth
	v1.GET("auth/user-info", authHandler.UserInfo)             // 获取登录账号信息
	v1.GET("auth/menus", authHandler.Menus)                    // 获取登录账号菜单
	v1.PUT("auth/change-password", authHandler.ChangePassword) // 修改密码
	v1.PUT("auth/switch-role", authHandler.SwitchRole)         // 切换角色

	// Security (TOTP)
	securityHandler := handlerauth.NewSecurityHandler(store.S)
	securityGroup := v1.Group("auth/security")
	{
		securityGroup.GET("/totp/status", securityHandler.GetTOTPStatus)
		securityGroup.POST("/totp/setup", securityHandler.SetupTOTP)
		securityGroup.POST("/totp/enable", securityHandler.EnableTOTP)
		securityGroup.POST("/totp/verify", securityHandler.VerifyTOTP)
		securityGroup.POST("/totp/disable", securityHandler.DisableTOTP)
	}

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
	v1.PUT("admins/:name/reset-totp", adminHandler.ResetTOTP)           // 重置 TOTP

	// Role
	roleHandler := system.NewRoleHandler(store.S, policyAuthz)
	v1.GET("roles", roleHandler.List)
	v1.POST("roles", roleHandler.Create)
	v1.GET("roles/:name", roleHandler.Get)
	v1.PUT("roles/:name", roleHandler.Update)
	v1.DELETE("roles/:name", roleHandler.Delete)
	v1.PUT("roles/:name/apis", roleHandler.SetApis)     // 设置权限（casbin）
	v1.GET("roles/:name/apis", roleHandler.GetApiIDs)   // 获取权限 ID 集合（casbin）
	v1.PUT("roles/:name/menus", roleHandler.SetMenus)   // 设置菜单权限
	v1.GET("roles/:name/menus", roleHandler.GetMenuIDs) // 获取菜单 ID 集合
	v1.GET("roles/all", roleHandler.All)

	// AI Agent
	aiAgentHandler := ai.NewAgentHandler(store.S)
	v1.GET("ai/agents", aiAgentHandler.List)
	v1.POST("ai/agents", aiAgentHandler.Create)
	v1.GET("ai/agents/:id", aiAgentHandler.Get)
	v1.PUT("ai/agents/:id", aiAgentHandler.Update)
	v1.DELETE("ai/agents/:id", aiAgentHandler.Delete)

	// AI Provider
	aiProviderHandler := ai.NewProviderHandler(store.S)
	v1.GET("ai/providers", aiProviderHandler.List)
	v1.GET("ai/providers/:id", aiProviderHandler.Get)
	v1.PUT("ai/providers/:id", aiProviderHandler.Update)

	// AI Model
	aiModelHandler := ai.NewModelHandler(store.S)
	v1.GET("ai/models", aiModelHandler.List)
	v1.POST("ai/models", aiModelHandler.Create)
	v1.GET("ai/models/:id", aiModelHandler.Get)
	v1.PUT("ai/models/:id", aiModelHandler.Update)
	v1.DELETE("ai/models/:id", aiModelHandler.Delete)

	// AI Quota
	aiQuotaHandler := ai.NewQuotaHandler(store.S)
	v1.GET("ai/quotas", aiQuotaHandler.List)
	v1.GET("ai/quotas/:uid", aiQuotaHandler.Get)
	v1.PUT("ai/quotas/:uid", aiQuotaHandler.Update)
	v1.POST("ai/quotas/:uid/reset-daily", aiQuotaHandler.ResetDailyTokens)

	// AI Health
	registry := aipkg.GetRegistry()
	aiHealthHandler := ai.NewHealthHandler(registry)
	v1.GET("ai/health", aiHealthHandler.GetHealthStatus)

	// API
	apiHandler := system.NewApiHandler(store.S, policyAuthz)
	v1.GET("apis", apiHandler.List)
	v1.GET("apis/all", apiHandler.All)
	v1.POST("apis", apiHandler.Create)
	v1.GET("apis/:id", apiHandler.Get)
	v1.PUT("apis/:id", apiHandler.Update)
	v1.DELETE("apis/:id", apiHandler.Delete)
	v1.GET("apis/tree", apiHandler.Tree)

	// Menu
	menuHandler := system.NewMenuHandler(store.S, policyAuthz)
	v1.GET("menus", menuHandler.List)
	v1.POST("menus", menuHandler.Create)
	v1.GET("menus/:id", menuHandler.Get)
	v1.PUT("menus/:id", menuHandler.Update)
	v1.DELETE("menus/:id", menuHandler.Delete)
	v1.GET("menus/tree", menuHandler.Tree)
	v1.POST("menus/:id/toggle-hidden", menuHandler.ToggleHidden)

	// App Version
	appVersionHandler := config.NewAppVersionHandler(store.S, policyAuthz)
	v1.GET("cfg/apps", appVersionHandler.List)
	v1.POST("cfg/apps", appVersionHandler.Create)
	v1.GET("cfg/apps/:id", appVersionHandler.Get)
	v1.PUT("cfg/apps/:id", appVersionHandler.Update)
	v1.DELETE("cfg/apps/:id", appVersionHandler.Delete)

	// Sys config
	configHandler := config.NewConfigHandler(store.S, policyAuthz)
	v1.GET("cfg/configs", configHandler.List)
	v1.POST("cfg/configs", configHandler.Create)
	v1.GET("cfg/configs/:id", configHandler.Get)
	v1.PUT("cfg/configs/:id", configHandler.Update)
	v1.DELETE("cfg/configs/:id", configHandler.Delete)

	// User
	userHandler := user.NewUserHandler(store.S, policyAuthz)
	v1.GET("users", userHandler.List)                                 // 列出用户列表，只有 root 用户才能访问
	v1.POST("users", userHandler.Create)                              // 创建用户
	v1.GET("users/:name", userHandler.Get)                            // 获取用户详情
	v1.PUT("users/:name", userHandler.Update)                         // 更新用户
	v1.DELETE("users/:name", userHandler.Delete)                      // 删除用户
	v1.PUT("users/:name/change-password", userHandler.ChangePassword) // 修改用户密码

	// App
	appHandler := app.NewAppHandler(store.S, policyAuthz)
	v1.GET("apps", appHandler.List)
	v1.POST("apps", appHandler.Create)
	v1.GET("apps/:appid", appHandler.Get)
	v1.PUT("apps/:appid", appHandler.Update)
	v1.DELETE("apps/:appid", appHandler.Delete)

	// Api keys
	apiKeyHandler := app.NewApiKeyHandler(store.S, policyAuthz)
	v1.GET("api-keys", apiKeyHandler.List)
	v1.POST("api-keys", apiKeyHandler.Create)
	v1.GET("api-keys/:id", apiKeyHandler.Get)
	v1.PUT("api-keys/:id", apiKeyHandler.Update)
	v1.DELETE("api-keys/:id", apiKeyHandler.Delete)

	// Announcements
	announcementHandler := notification.NewAnnouncementHandler(store.S, policyAuthz)
	v1.GET("announcements", announcementHandler.List)
	v1.POST("announcements", announcementHandler.Create)
	v1.GET("announcements/:uuid", announcementHandler.Get)
	v1.PUT("announcements/:uuid", announcementHandler.Update)
	v1.DELETE("announcements/:uuid", announcementHandler.Delete)
	v1.POST("announcements/:uuid/publish", announcementHandler.Publish)
	v1.POST("announcements/:uuid/schedule", announcementHandler.Schedule)
	v1.POST("announcements/:uuid/cancel", announcementHandler.Cancel)
}
