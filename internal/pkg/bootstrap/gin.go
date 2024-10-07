package bootstrap

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	ginprom "github.com/zsais/go-gin-prometheus"

	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/http/middleware"
)

func InitGin() *gin.Engine {
	gin.SetMode(facade.Config.Server.Mode)

	g := gin.New()

	// Register global middlewares
	registerGlobalMiddleWare(g)

	// Register static file server
	registerStaticFileServer(g)

	// Metrics
	if facade.Config.Feature.Metrics {
		prometheus := ginprom.NewPrometheus("gin")
		prometheus.Use(g)
	}

	// Profiling
	if facade.Config.Feature.Profiling {
		registerProfiling(g)
	}

	return g
}

func registerGlobalMiddleWare(g *gin.Engine) {
	g.Use(
		gin.Recovery(),
		middleware.NoCache,
		middleware.Cors,
		middleware.Secure,
		middleware.ForceUserAgent,
		middleware.RequestID(),
		middleware.Context(),
		middleware.LimitWrite("1-S"), // 限制写操作，每秒 1 次
		middleware.LimitIP("20-S"),   // 限制 IP 请求，每秒 20 次
		middleware.Logger(),
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

func registerProfiling(g *gin.Engine) {
	p := g.Group("system")
	p.Use(middleware.Debug())

	pprof.RouteRegister(p, "debug/pprof")
}
