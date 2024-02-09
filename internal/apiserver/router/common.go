package router

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/http/controller/v1/common"
	"bingo/internal/apiserver/http/controller/v1/file"
	"bingo/internal/apiserver/http/middleware"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
)

func MapCommonRouters(g *gin.Engine) {
	// 注册 404 Handler.
	g.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, errno.ErrResourceNotFound, nil)
	})

	// Common
	commonController := common.NewCommonController()
	g.GET("/healthz", commonController.Healthz)
	g.GET("/version", commonController.Version)

	// v1 group
	v1 := g.Group("/v1")

	// Upload
	fileController := file.NewFileController(nil, nil)
	v1.POST("file/upload", middleware.Authn(), fileController.Upload)
}
