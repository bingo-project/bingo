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

type AppController struct {
	a *auth.Authz
	b biz.IBiz
}

func NewAppController(ds store.IStore, a *auth.Authz) *AppController {
	return &AppController{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List apps
// @Security   Bearer
// @Tags       System.App
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListAppRequest	 true  "Param"
// @Success	   200		{object}	v1.ListAppResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/cfg/apps [GET].
func (ctrl *AppController) List(c *gin.Context) {
	log.C(c).Infow("List app function called")

	var req v1.ListAppRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	resp, err := ctrl.b.Apps().List(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Create
// @Summary    Create app
// @Security   Bearer
// @Tags       System.App
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateAppRequest	 true  "Param"
// @Success	   200		{object}	v1.AppInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/cfg/apps [POST].
func (ctrl *AppController) Create(c *gin.Context) {
	log.C(c).Infow("Create app function called")

	var req v1.CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create app
	resp, err := ctrl.b.Apps().Create(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Get
// @Summary    Get app info
// @Security   Bearer
// @Tags       System.App
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.AppInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/cfg/apps/{id} [GET].
func (ctrl *AppController) Get(c *gin.Context) {
	log.C(c).Infow("Get app function called")

	ID := cast.ToUint(c.Param("id"))
	app, err := ctrl.b.Apps().Get(c, ID)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, app)
}

// Update
// @Summary    Update app info
// @Security   Bearer
// @Tags       System.App
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdateAppRequest	 true  "Param"
// @Success	   200		{object}	v1.AppInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/cfg/apps/{id} [PUT].
func (ctrl *AppController) Update(c *gin.Context) {
	log.C(c).Infow("Update app function called")

	var req v1.UpdateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.Apps().Update(c, ID, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete
// @Summary    Delete app
// @Security   Bearer
// @Tags       System.App
// @Accept     application/json
// @Produce    json
// @Param      id	    path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/cfg/apps/{id} [DELETE].
func (ctrl *AppController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete app function called")

	ID := cast.ToUint(c.Param("id"))
	if err := ctrl.b.Apps().Delete(c, ID); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
