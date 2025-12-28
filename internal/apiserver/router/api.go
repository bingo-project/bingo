package router

import (
	"github.com/gin-gonic/gin"

	bizauth "github.com/bingo-project/bingo/internal/apiserver/biz/auth"
	authhandler "github.com/bingo-project/bingo/internal/apiserver/handler/http/auth"
	ntfhandler "github.com/bingo-project/bingo/internal/apiserver/handler/http/notification"
	"github.com/bingo-project/bingo/internal/apiserver/middleware"
	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

func MapApiRouters(g *gin.Engine) {
	// v1 group
	v1 := g.Group("/v1")
	v1.Use(middleware.Lang())
	v1.Use(middleware.Maintenance())

	authHandler := authhandler.NewAuthHandler(store.S, nil)

	// Auth routes
	authGroup := v1.Group("/auth")
	{
		// 公开接口
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/code", authHandler.SendCode)
		authGroup.POST("/reset-password", authHandler.ResetPassword)

		// Login by Address
		authGroup.GET("/nonce", authHandler.Nonce)
		authGroup.POST("/login/address", authHandler.LoginByAddress)

		// OAuth
		authGroup.GET("/providers", authHandler.Providers)
		authGroup.GET("/login/:provider", authHandler.GetAuthCode)
		authGroup.POST("/login/:provider", authHandler.LoginByProvider)
	}

	// Authentication middleware
	loader := bizauth.NewUserLoader(store.S)
	authn := auth.New(loader)
	v1.Use(auth.Middleware(authn))

	// 需要登录的接口
	authAuthed := v1.Group("/auth")
	{
		authAuthed.GET("/user-info", authHandler.UserInfo)
		authAuthed.PUT("/user", authHandler.UpdateProfile)
		authAuthed.PUT("/change-password", authHandler.ChangePassword)

		// 社交账号管理
		authAuthed.GET("/bindings", authHandler.ListBindings)
		authAuthed.POST("/bindings/:provider", authHandler.BindProvider)
		authAuthed.DELETE("/bindings/:provider", authHandler.Unbind)
	}

	// Security settings
	securityHandler := authhandler.NewSecurityHandler(store.S)
	securityGroup := v1.Group("/auth/security")
	{
		securityGroup.GET("/status", securityHandler.GetSecurityStatus)

		// Pay password
		securityGroup.PUT("/pay-password", securityHandler.SetPayPassword)
		securityGroup.POST("/verify-pay-password", securityHandler.VerifyPayPassword)

		// TOTP
		securityGroup.GET("/totp/status", securityHandler.GetTOTPStatus)
		securityGroup.POST("/totp/setup", securityHandler.SetupTOTP)
		securityGroup.POST("/totp/enable", securityHandler.EnableTOTP)
		securityGroup.POST("/totp/verify", securityHandler.VerifyTOTP)
		securityGroup.POST("/totp/disable", securityHandler.DisableTOTP)
	}

	// Notification routes
	ntfHandler := ntfhandler.NewNotificationHandler(store.S)
	prefHandler := ntfhandler.NewPreferenceHandler(store.S)

	ntf := v1.Group("/notifications")
	{
		ntf.GET("", ntfHandler.List)
		ntf.GET("/unread-count", ntfHandler.UnreadCount)
		ntf.PUT("/:uuid/read", ntfHandler.MarkAsRead)
		ntf.PUT("/read-all", ntfHandler.MarkAllAsRead)
		ntf.DELETE("/:uuid", ntfHandler.Delete)
		ntf.GET("/preferences", prefHandler.Get)
		ntf.PUT("/preferences", prefHandler.Update)
	}
}
