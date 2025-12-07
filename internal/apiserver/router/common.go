package router

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/handler/http/common"
	"bingo/internal/apiserver/handler/http/file"
	"bingo/internal/apiserver/middleware"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/store"
)

func MapCommonRouters(g *gin.Engine) {
	// 注册 404 Handler.
	g.NoRoute(func(c *gin.Context) {
		core.Response(c, nil, errno.ErrNotFound)
	})

	cm := g.Group("/")
	cm.Use(middleware.Maintenance())

	// Common
	commonHandler := common.NewCommonHandler(store.S)
	cm.GET("/healthz", commonHandler.Healthz)
	cm.GET("/version", commonHandler.Version)

	// v1 group
	v1 := g.Group("/v1")

	// Upload
	fileHandler := file.NewFileHandler(nil, nil)
	v1.POST("file/upload", middleware.Authn(), fileHandler.Upload)
}
