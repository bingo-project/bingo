package common

import (
	"github.com/bingo-project/component-base/version"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/biz"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1"
)

type CommonHandler struct {
	b biz.IBiz
}

func NewCommonHandler(ds store.IStore) *CommonHandler {
	return &CommonHandler{
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
func (ctrl *CommonHandler) Healthz(c *gin.Context) {
	log.C(c).Infow("Healthz function called")

	status, err := ctrl.b.Servers().Status(c)
	if err != nil {
		return
	}

	data := &v1.HealthzResponse{Status: status}

	core.Response(c, data, nil)
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
func (ctrl *CommonHandler) Version(c *gin.Context) {
	log.C(c).Infow("Version function called")

	core.Response(c, version.Get(), nil)
}
