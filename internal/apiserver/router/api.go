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

	// Login
	v1.POST("auth/code/email", authHandler.SendEmailCode)
	v1.POST("auth/register", authHandler.Register)
	v1.POST("auth/login", authHandler.Login)

	// Login by Address
	v1.GET("auth/nonce", authHandler.Nonce)
	v1.POST("auth/login/address", authHandler.LoginByAddress)

	// Login by Third Party
	v1.GET("auth/providers", authHandler.Providers)
	v1.GET("auth/login/:provider", authHandler.GetAuthCode)
	v1.POST("auth/login/:provider", authHandler.LoginByProvider)

	// Authentication middleware
	loader := bizauth.NewUserLoader(store.S)
	authn := auth.New(loader)
	v1.Use(auth.Middleware(authn))

	// Auth
	v1.GET("auth/user-info", authHandler.UserInfo)             // 获取登录账号信息
	v1.PUT("auth/change-password", authHandler.ChangePassword) // 修改用户密码
	v1.POST("auth/bind/:provider", authHandler.BindProvider)
}
