package common

import (
	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/version"
	"github.com/gin-gonic/gin"

	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/pkg/core"
)

type CommonController struct{}

// NewCommonController 创建一个 common controller.
func NewCommonController() *CommonController {
	return &CommonController{}
}

// Healthz
// @Summary    Heath check
// @Tags       Common
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	v1.HealthzResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /healthz  [GET].
func (ctrl *CommonController) Healthz(c *gin.Context) {
	log.C(c).Infow("Healthz function called")

	data := &v1.HealthzResponse{Status: "ok"}

	core.WriteResponse(c, nil, data)
}

// Version
// @Summary    Get App Version
// @Tags       Common
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	version.Info
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /version  [GET].
func (ctrl *CommonController) Version(c *gin.Context) {
	log.C(c).Infow("Version function called")

	core.WriteResponse(c, nil, version.Get())
}
