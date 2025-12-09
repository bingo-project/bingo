// ABOUTME: HTTP server initialization for apiserver.
// ABOUTME: Configures Gin engine with routes and middleware.

package apiserver

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/apiserver/router"
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

	// Common router
	router.MapCommonRouters(g)

	// Api
	router.MapApiRouters(g)

	return g
}
