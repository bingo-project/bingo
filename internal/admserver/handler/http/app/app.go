package app

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/admserver/biz"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1"
	"bingo/pkg/auth"
)

type AppHandler struct {
	a *auth.Authz
	b biz.IBiz
}

func NewAppHandler(ds store.IStore, a *auth.Authz) *AppHandler {
	return &AppHandler{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List apps
// @Security   Bearer
// @Tags       App
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListAppRequest	 true  "Param"
// @Success	   200		{object}	v1.ListAppResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apps [GET].
func (ctrl *AppHandler) List(c *gin.Context) {
	log.C(c).Infow("List app function called")

	var req v1.ListAppRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	resp, err := ctrl.b.Apps().List(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Create
// @Summary    Create app
// @Security   Bearer
// @Tags       App
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateAppRequest	 true  "Param"
// @Success	   200		{object}	v1.AppInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apps [POST].
func (ctrl *AppHandler) Create(c *gin.Context) {
	log.C(c).Infow("Create app function called")

	var req v1.CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	// Create app
	resp, err := ctrl.b.Apps().Create(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Get
// @Summary    Get app info
// @Security   Bearer
// @Tags       App
// @Accept     application/json
// @Produce    json
// @Param      appid	 path	    string            		 true  "ID"
// @Success	   200		{object}	v1.AppInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apps/{appid} [GET].
func (ctrl *AppHandler) Get(c *gin.Context) {
	log.C(c).Infow("Get app function called")

	appID := c.Param("appid")
	app, err := ctrl.b.Apps().Get(c, appID)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, app, nil)
}

// Update
// @Summary    Update app info
// @Security   Bearer
// @Tags       App
// @Accept     application/json
// @Produce    json
// @Param      appid	 path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdateAppRequest	 true  "Param"
// @Success	   200		{object}	v1.AppInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apps/{appid} [PUT].
func (ctrl *AppHandler) Update(c *gin.Context) {
	log.C(c).Infow("Update app function called")

	var req v1.UpdateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	appID := c.Param("appid")
	resp, err := ctrl.b.Apps().Update(c, appID, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Delete
// @Summary    Delete app
// @Security   Bearer
// @Tags       App
// @Accept     application/json
// @Produce    json
// @Param      appid	 path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apps/{appid} [DELETE].
func (ctrl *AppHandler) Delete(c *gin.Context) {
	log.C(c).Infow("Delete app function called")

	appID := c.Param("appid")
	if err := ctrl.b.Apps().Delete(c, appID); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}
