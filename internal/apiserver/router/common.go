package router

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/http/controller/v1/common"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
)

func MapCommonRouters(g *gin.Engine) {
	// 注册 404 Handler.
	g.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, errno.ErrResourceNotFound, nil)
	})

	// Healthz
	commonController := common.NewCommonController()
	g.GET("/healthz", commonController.Healthz)
}
