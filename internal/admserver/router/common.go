package router

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/admserver/handler/http/common"
	"bingo/internal/admserver/handler/http/file"
	"bingo/internal/admserver/middleware"
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
	commonController := common.NewCommonController(store.S)
	cm.GET("/healthz", commonController.Healthz)
	cm.GET("/version", commonController.Version)

	// v1 group
	v1 := g.Group("/v1")

	// Upload
	fileController := file.NewFileController(nil, nil)
	v1.POST("file/upload", middleware.Authn(), fileController.Upload)
}
