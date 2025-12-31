// ABOUTME: HTTP server initialization for apiserver.
// ABOUTME: Configures Gin engine with routes and middleware.

package apiserver

import (
	"github.com/gin-gonic/gin"

	bizauth "github.com/bingo-project/bingo/internal/apiserver/biz/auth"
	"github.com/bingo-project/bingo/internal/apiserver/middleware"
	"github.com/bingo-project/bingo/internal/apiserver/router"
	"github.com/bingo-project/bingo/internal/pkg/ai"
	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/bootstrap"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	httpmw "github.com/bingo-project/bingo/internal/pkg/middleware/http"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

// initGinEngine initializes the Gin engine with routes.
func initGinEngine() *gin.Engine {
	g := bootstrap.InitGin()

	// Swagger
	if facade.Config.Feature.ApiDoc {
		router.MapSwagRouters(g)
	}

	// Common router
	router.MapCommonRouters(g)

	// Api
	router.MapApiRouters(g)

	// AI Chat routes (use global registry)
	if registry := ai.GetRegistry(); registry != nil {
		v1 := g.Group("/v1")
		v1.Use(middleware.Lang())
		v1.Use(middleware.Maintenance())

		loader := bizauth.NewUserLoader(store.S)
		authn := auth.New(loader)
		v1.Use(auth.Middleware(authn))

		// Apply AI rate limiter (RPM)
		rpm := facade.Config.AI.Quota.DefaultRPM
		if rpm <= 0 {
			rpm = 20 // fallback default
		}
		v1.Use(httpmw.AILimiter(rpm))

		router.MapAiRouters(v1, registry)
	}

	return g
}
