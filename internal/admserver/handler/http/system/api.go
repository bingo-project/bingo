package system

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/admserver/biz"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/store"
	"bingo/pkg/api/apiserver/v1"
	"bingo/pkg/auth"
)

type ApiController struct {
	a *auth.Authz
	b biz.IBiz
}

func NewApiController(ds store.IStore, a *auth.Authz) *ApiController {
	return &ApiController{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List apis
// @Security   Bearer
// @Tags       Api
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListApiRequest	 true  "Param"
// @Success	   200		{object}	v1.ListApiResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apis [GET].
func (ctrl *ApiController) List(c *gin.Context) {
	log.C(c).Infow("List api function called")

	var req v1.ListApiRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	resp, err := ctrl.b.Apis().List(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Create
// @Summary    Create a api
// @Security   Bearer
// @Tags       Api
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateApiRequest	 true  "Param"
// @Success	   200		{object}	v1.ApiInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apis [POST].
func (ctrl *ApiController) Create(c *gin.Context) {
	log.C(c).Infow("Create api function called")

	var req v1.CreateApiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create api
	resp, err := ctrl.b.Apis().Create(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Get
// @Summary    Get api info
// @Security   Bearer
// @Tags       Api
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.ApiInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apis/{id} [GET].
func (ctrl *ApiController) Get(c *gin.Context) {
	log.C(c).Infow("Get api function called")

	ID := cast.ToUint(c.Param("id"))
	api, err := ctrl.b.Apis().Get(c, ID)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, api)
}

// Update
// @Summary    Update api info
// @Security   Bearer
// @Tags       Api
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdateApiRequest	 true  "Param"
// @Success	   200		{object}	v1.ApiInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apis/{id} [PUT].
func (ctrl *ApiController) Update(c *gin.Context) {
	log.C(c).Infow("Update api function called")

	var req v1.UpdateApiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.Apis().Update(c, ID, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete
// @Summary    Delete api
// @Security   Bearer
// @Tags       Api
// @Accept     application/json
// @Produce    json
// @Param      id	    path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apis/{id} [DELETE].
func (ctrl *ApiController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete api function called")

	ID := cast.ToUint(c.Param("id"))
	if err := ctrl.b.Apis().Delete(c, ID); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}

// All
// @Summary    All apis
// @Security   Bearer
// @Tags       Api
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	[]v1.ApiInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apis/all [GET].
func (ctrl *ApiController) All(c *gin.Context) {
	log.C(c).Infow("All api function called")

	resp, err := ctrl.b.Apis().All(c)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Tree
// @Summary    API Tree
// @Security   Bearer
// @Tags       Api
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	[]v1.GroupApiResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/apis/tree [GET].
func (ctrl *ApiController) Tree(c *gin.Context) {
	log.C(c).Infow("Tree api function called")

	resp, err := ctrl.b.Apis().Tree(c)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}
