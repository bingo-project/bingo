// ABOUTME: HTTP server initialization for admserver.
// ABOUTME: Configures Gin engine with routes and middleware.

package admserver

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/admserver/router"
	"github.com/bingo-project/bingo/internal/pkg/bootstrap"
	"github.com/bingo-project/bingo/internal/pkg/facade"
)

// initGinEngine initializes the Gin engine with routes.
func initGinEngine() *gin.Engine {
	g := bootstrap.InitGin()

	// Swagger
	if facade.Config.Feature.ApiDoc {
		router.MapSwagRouters(g)
	}

	// Queue dashboard
	if facade.Config.Feature.QueueDash {
		router.MapQueueRouters(g)
	}

	// Common router
	router.MapCommonRouters(g)

	// System
	router.MapApiRouters(g)

	return g
}
