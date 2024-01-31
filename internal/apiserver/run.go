package apiserver

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/bootstrap"
	"bingo/internal/apiserver/facade"
	"bingo/internal/apiserver/http/middleware"
	"bingo/internal/apiserver/router"
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

	// Register static file server
	registerStaticFileServer(g)

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

	// Init System API
	router.InitSystemAPI(g)

	return g
}

// Register global middlewares.
func registerGlobalMiddleWare(g *gin.Engine) {
	g.Use(
		gin.Recovery(),
		middleware.NoCache,
		middleware.Cors,
		middleware.Secure,
		middleware.ForceUserAgent,
		middleware.RequestID(),
		middleware.LimitWrite("1-S"), // 限制写操作，每秒 1 次
		middleware.LimitIP("20-S"),   // 限制 IP 请求，每秒 20 次
	)
}

// Register static file server.
func registerStaticFileServer(g *gin.Engine) {
	storage := g.Group("storage")

	// Upload for user
	storage.Static("upload", "./storage/public/upload")

	// Authentication for secure file.
	storage.Use(middleware.Authn())
	storage.Static("log", "./storage/log")
}
