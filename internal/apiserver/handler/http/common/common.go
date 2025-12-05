package common

import (
	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/version"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/biz"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/store"
	"bingo/pkg/api/apiserver/v1"
)

type CommonController struct {
	b biz.IBiz
}

// NewCommonController 创建一个 common controller.
func NewCommonController(ds store.IStore) *CommonController {
	return &CommonController{
		b: biz.NewBiz(ds),
	}
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

	status, err := ctrl.b.Servers().Status(c)
	if err != nil {
		return
	}

	data := &v1.HealthzResponse{Status: status}

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
