package router

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/controller/v1/common"
)

func MapCommonRouters(r *gin.Engine) {
	/**
	|--------------------------------------------------------------------------
	| Common
	|--------------------------------------------------------------------------
	|
	| Here is where you can register API routes for your application. These
	| routes are loaded by the RouteServiceProvider within a group which
	| is assigned the "api" middleware group. Enjoy building your API!
	|
	*/

	// controllers
	commonController := common.NewCommonController()

	// 注册 /healthz handler.
	r.GET("/healthz", commonController.Healthz)
}
