package router

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/controller/v1/common"
)

func MapCommonRouters(r *gin.Engine) {
	// controllers
	commonController := common.NewCommonController()

	// 注册 /healthz handler.
	r.GET("/healthz", commonController.Healthz)
}
