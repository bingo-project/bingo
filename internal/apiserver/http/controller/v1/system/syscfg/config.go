package syscfg

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/apiserver/biz"
	v1 "bingo/internal/apiserver/http/request/v1/syscfg"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/pkg/auth"
)

type ConfigController struct {
	a *auth.Authz
	b biz.IBiz
}

func NewConfigController(ds store.IStore, a *auth.Authz) *ConfigController {
	return &ConfigController{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List configs
// @Security   Bearer
// @Tags       System.Config
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListConfigRequest	 true  "Param"
// @Success	   200		{object}	v1.ListConfigResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/cfg/configs [GET]
func (ctrl *ConfigController) List(c *gin.Context) {
	log.C(c).Infow("List config function called")

	var req v1.ListConfigRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	resp, err := ctrl.b.Configs().List(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Create
// @Summary    Create config
// @Security   Bearer
// @Tags       System.Config
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateConfigRequest	 true  "Param"
// @Success	   200		{object}	v1.ConfigInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/cfg/configs [POST]
func (ctrl *ConfigController) Create(c *gin.Context) {
	log.C(c).Infow("Create config function called")

	var req v1.CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create config
	resp, err := ctrl.b.Configs().Create(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Get
// @Summary    Get config info
// @Security   Bearer
// @Tags       System.Config
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.ConfigInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/cfg/configs/{id} [GET]
func (ctrl *ConfigController) Get(c *gin.Context) {
	log.C(c).Infow("Get config function called")

	ID := cast.ToUint(c.Param("id"))
	config, err := ctrl.b.Configs().Get(c, ID)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, config)
}

// Update
// @Summary    Update config info
// @Security   Bearer
// @Tags       System.Config
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdateConfigRequest	 true  "Param"
// @Success	   200		{object}	v1.ConfigInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/cfg/configs/{id} [PUT]
func (ctrl *ConfigController) Update(c *gin.Context) {
	log.C(c).Infow("Update config function called")

	var req v1.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.Configs().Update(c, ID, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete
// @Summary    Delete config
// @Security   Bearer
// @Tags       System.Config
// @Accept     application/json
// @Produce    json
// @Param      id	    path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/cfg/configs/{id} [DELETE]
func (ctrl *ConfigController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete config function called")

	ID := cast.ToUint(c.Param("id"))
	if err := ctrl.b.Configs().Delete(c, ID); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
