package apiserver

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/bootstrap"
	"bingo/internal/apiserver/facade"
	"bingo/internal/apiserver/router"
	"bingo/internal/pkg/middleware"
)

// run 函数是实际的业务代码入口函数.
func run() error {
	bootstrap.Boot()

	g := initRouter()

	// 创建并运行 HTTP 服务器
	return startInsecureServer(g)
}

func initRouter() *gin.Engine {
	gin.SetMode(facade.Config.Server.Mode)

	g := gin.New()

	// Register global middlewares
	registerGlobalMiddleWare(g)

	// Swagger
	if facade.Config.Feature.ApiDoc {
		router.MapSwagRouters(g)
	}

	// Common router
	router.MapCommonRouters(g)

	// Api
	router.MapApiRouters(g)

	// System
	router.MapSystemRouters(g)

	return g
}

// Register global middlewares
func registerGlobalMiddleWare(g *gin.Engine) {
	g.Use(
		gin.Recovery(),
		middleware.NoCache,
		middleware.Cors,
		middleware.Secure,
		middleware.ForceUserAgent,
		middleware.RequestID(),
	)
}
