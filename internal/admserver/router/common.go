package router

import (
	"github.com/gin-gonic/gin"

	bizauth "bingo/internal/admserver/biz/auth"
	"bingo/internal/admserver/handler/http/common"
	"bingo/internal/admserver/handler/http/file"
	"bingo/internal/admserver/middleware"
	"bingo/internal/pkg/auth"
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
	loader := bizauth.NewAdminLoader(store.S)
	authn := auth.New(loader)
	fileHandler := file.NewFileHandler(nil, nil)
	v1.POST("file/upload", auth.Middleware(authn), fileHandler.Upload)
}
