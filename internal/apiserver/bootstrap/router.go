package bootstrap

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/config"
	"bingo/internal/apiserver/router"
	"bingo/internal/pkg/middleware"
)

func InitRouter() *gin.Engine {
	gin.SetMode(config.Cfg.Server.Mode)

	g := gin.New()

	// Register global middlewares
	registerGlobalMiddleWare(g)

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
