package app

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/apiserver/biz"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/pkg/api/apiserver/v1"
	"bingo/pkg/auth"
)

type ApiKeyController struct {
	a *auth.Authz
	b biz.IBiz
}

func NewApiKeyController(ds store.IStore, a *auth.Authz) *ApiKeyController {
	return &ApiKeyController{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List apiKeys
// @Security   Bearer
// @Tags       System.App
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListApiKeyRequest	 true  "Param"
// @Success	   200		{object}	v1.ListApiKeyResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/api-keys [GET]
func (ctrl *ApiKeyController) List(c *gin.Context) {
	log.C(c).Infow("List apiKey function called")

	var req v1.ListApiKeyRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	resp, err := ctrl.b.ApiKeys().List(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Create
// @Summary    Create apiKey
// @Security   Bearer
// @Tags       System.App
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateApiKeyRequest	 true  "Param"
// @Success	   200		{object}	v1.ApiKeyInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/api-keys [POST]
func (ctrl *ApiKeyController) Create(c *gin.Context) {
	log.C(c).Infow("Create apiKey function called")

	var req v1.CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create apiKey
	resp, err := ctrl.b.ApiKeys().Create(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Get
// @Summary    Get apiKey info
// @Security   Bearer
// @Tags       System.App
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.ApiKeyInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/api-keys/{id} [GET]
func (ctrl *ApiKeyController) Get(c *gin.Context) {
	log.C(c).Infow("Get apiKey function called")

	ID := cast.ToUint(c.Param("id"))
	apiKey, err := ctrl.b.ApiKeys().Get(c, ID)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, apiKey)
}

// Update
// @Summary    Update apiKey info
// @Security   Bearer
// @Tags       System.App
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdateApiKeyRequest	 true  "Param"
// @Success	   200		{object}	v1.ApiKeyInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/api-keys/{id} [PUT]
func (ctrl *ApiKeyController) Update(c *gin.Context) {
	log.C(c).Infow("Update apiKey function called")

	var req v1.UpdateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.ApiKeys().Update(c, ID, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete
// @Summary    Delete apiKey
// @Security   Bearer
// @Tags       System.App
// @Accept     application/json
// @Produce    json
// @Param      id	    path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/api-keys/{id} [DELETE]
func (ctrl *ApiKeyController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete apiKey function called")

	ID := cast.ToUint(c.Param("id"))
	if err := ctrl.b.ApiKeys().Delete(c, ID); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
