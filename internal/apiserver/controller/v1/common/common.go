package common

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/core"
	v1 "bingo/pkg/api/bingo/v1"
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
