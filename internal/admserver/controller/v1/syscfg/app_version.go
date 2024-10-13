package syscfg

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/admserver/biz"
	"bingo/internal/admserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	v1 "bingo/pkg/api/apiserver/v1/syscfg"
	"bingo/pkg/auth"
)

type AppVersionController struct {
	a *auth.Authz
	b biz.IBiz
}

func NewAppVersionController(ds store.IStore, a *auth.Authz) *AppVersionController {
	return &AppVersionController{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List apps
// @Security   Bearer
// @Tags       Config
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListAppVersionRequest	 true  "Param"
// @Success	   200		{object}	v1.ListAppVersionResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/cfg/apps [GET].
func (ctrl *AppVersionController) List(c *gin.Context) {
	log.C(c).Infow("List app function called")

	var req v1.ListAppVersionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	resp, err := ctrl.b.AppVersions().List(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Create
// @Summary    Create app
// @Security   Bearer
// @Tags       Config
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateAppVersionRequest	 true  "Param"
// @Success	   200		{object}	v1.AppVersionInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/cfg/apps [POST].
func (ctrl *AppVersionController) Create(c *gin.Context) {
	log.C(c).Infow("Create app function called")

	var req v1.CreateAppVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create app
	resp, err := ctrl.b.AppVersions().Create(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Get
// @Summary    Get app info
// @Security   Bearer
// @Tags       Config
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.AppVersionInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/cfg/apps/{id} [GET].
func (ctrl *AppVersionController) Get(c *gin.Context) {
	log.C(c).Infow("Get app function called")

	ID := cast.ToUint(c.Param("id"))
	app, err := ctrl.b.AppVersions().Get(c, ID)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, app)
}

// Update
// @Summary    Update app info
// @Security   Bearer
// @Tags       Config
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdateAppVersionRequest	 true  "Param"
// @Success	   200		{object}	v1.AppVersionInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/cfg/apps/{id} [PUT].
func (ctrl *AppVersionController) Update(c *gin.Context) {
	log.C(c).Infow("Update app function called")

	var req v1.UpdateAppVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.AppVersions().Update(c, ID, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete
// @Summary    Delete app
// @Security   Bearer
// @Tags       Config
// @Accept     application/json
// @Produce    json
// @Param      id	    path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/cfg/apps/{id} [DELETE].
func (ctrl *AppVersionController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete app function called")

	ID := cast.ToUint(c.Param("id"))
	if err := ctrl.b.AppVersions().Delete(c, ID); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
