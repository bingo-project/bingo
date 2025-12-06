// ABOUTME: HTTP server initialization for apiserver.
// ABOUTME: Configures Gin engine with routes and middleware.

package apiserver

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/router"
	"bingo/internal/pkg/bootstrap"
	"bingo/internal/pkg/facade"
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
