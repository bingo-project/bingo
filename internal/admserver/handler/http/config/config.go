package config

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/admserver/biz"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1/syscfg"
	"bingo/pkg/auth"
)

type ConfigHandler struct {
	a *auth.Authz
	b biz.IBiz
}

func NewConfigHandler(ds store.IStore, a *auth.Authz) *ConfigHandler {
	return &ConfigHandler{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List configs
// @Security   Bearer
// @Tags       Config
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListConfigRequest	 true  "Param"
// @Success	   200		{object}	v1.ListConfigResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/cfg/configs [GET].
func (ctrl *ConfigHandler) List(c *gin.Context) {
	log.C(c).Infow("List config function called")

	var req v1.ListConfigRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	resp, err := ctrl.b.Configs().List(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Create
// @Summary    Create config
// @Security   Bearer
// @Tags       Config
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateConfigRequest	 true  "Param"
// @Success	   200		{object}	v1.ConfigInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/cfg/configs [POST].
func (ctrl *ConfigHandler) Create(c *gin.Context) {
	log.C(c).Infow("Create config function called")

	var req v1.CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	// Create config
	resp, err := ctrl.b.Configs().Create(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Get
// @Summary    Get config info
// @Security   Bearer
// @Tags       Config
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.ConfigInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/cfg/configs/{id} [GET].
func (ctrl *ConfigHandler) Get(c *gin.Context) {
	log.C(c).Infow("Get config function called")

	ID := cast.ToUint(c.Param("id"))
	config, err := ctrl.b.Configs().Get(c, ID)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, config, nil)
}

// Update
// @Summary    Update config info
// @Security   Bearer
// @Tags       Config
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdateConfigRequest	 true  "Param"
// @Success	   200		{object}	v1.ConfigInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/cfg/configs/{id} [PUT].
func (ctrl *ConfigHandler) Update(c *gin.Context) {
	log.C(c).Infow("Update config function called")

	var req v1.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.Configs().Update(c, ID, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Delete
// @Summary    Delete config
// @Security   Bearer
// @Tags       Config
// @Accept     application/json
// @Produce    json
// @Param      id	    path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/cfg/configs/{id} [DELETE].
func (ctrl *ConfigHandler) Delete(c *gin.Context) {
	log.C(c).Infow("Delete config function called")

	ID := cast.ToUint(c.Param("id"))
	if err := ctrl.b.Configs().Delete(c, ID); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}
