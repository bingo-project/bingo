package bootstrap

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	ginprom "github.com/zsais/go-gin-prometheus"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	middleware "github.com/bingo-project/bingo/internal/pkg/middleware/http"
)

func InitGin() *gin.Engine {
	gin.SetMode(facade.Config.HTTP.GinMode)

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
		middleware.ClientIP(),
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

	// Authentication for secure file (token validation only, no user loading).
	authn := auth.New(nil)
	storage.Use(auth.Middleware(authn))
	storage.Static("log", "./storage/log")
}

func registerProfiling(g *gin.Engine) {
	p := g.Group("system")
	p.Use(middleware.Debug())

	pprof.RouteRegister(p, "debug/pprof")
}

// InitGinForWebSocket creates a minimal Gin engine for WebSocket connections.
// WebSocket connections don't need rate limiting, User-Agent checks, or request logging.
func InitGinForWebSocket() *gin.Engine {
	gin.SetMode(facade.Config.HTTP.GinMode)

	g := gin.New()
	g.Use(
		gin.Recovery(),
		middleware.Cors,
		middleware.RequestID(),
	)

	return g
}
