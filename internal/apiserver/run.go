package apiserver

import (
	"github.com/bingo-project/component-base/web/token"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/config"
	"bingo/internal/apiserver/router"
	"bingo/internal/pkg/middleware"
)

// run 函数是实际的业务代码入口函数.
func run() error {
	// 初始化 store 层
	if err := InitStore(); err != nil {
		return err
	}

	// 初始化 cache
	if err := InitCache(); err != nil {
		return err
	}

	// 设置 token 包的签发密钥，用于 token 包 token 的签发和解析
	token.Init(config.Cfg.JWT.SecretKey, config.Cfg.JWT.TTL)

	// 设置 Gin 模式
	gin.SetMode(config.Cfg.Server.Mode)

	// 创建 Gin 引擎
	g := gin.New()

	// gin.Recovery() 中间件，用来捕获任何 panic，并恢复
	mws := []gin.HandlerFunc{gin.Recovery(), middleware.NoCache, middleware.Cors, middleware.Secure, middleware.RequestID()}

	g.Use(mws...)

	// Swagger
	if config.Cfg.Feature.ApiDoc {
		router.MapSwagRouters(g)
	}

	// Common router
	router.MapCommonRouters(g)

	// Api
	router.MapApiRouters(g)

	// System
	router.MapSystemRouters(g)

	// 创建并运行 HTTP 服务器
	return startInsecureServer(g)
}
