package router

import (
	"github.com/gin-gonic/gin"

	bizauth "github.com/bingo-project/bingo/internal/apiserver/biz/auth"
	authhandler "github.com/bingo-project/bingo/internal/apiserver/handler/http/auth"
	"github.com/bingo-project/bingo/internal/apiserver/middleware"
	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

func MapApiRouters(g *gin.Engine) {
	// v1 group
	v1 := g.Group("/v1")
	v1.Use(middleware.Maintenance())

	authHandler := authhandler.NewAuthHandler(store.S, nil)

	// Auth routes
	authGroup := v1.Group("/auth")
	{
		// 公开接口
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/code", authHandler.SendCode)
		authGroup.POST("/code/email", authHandler.SendEmailCode)
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
}
