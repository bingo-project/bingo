package app

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/admserver/biz"
	"bingo/internal/pkg/auth"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1"
)

type ApiKeyHandler struct {
	a *auth.Authorizer
	b biz.IBiz
}

func NewApiKeyHandler(ds store.IStore, a *auth.Authorizer) *ApiKeyHandler {
	return &ApiKeyHandler{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List apiKeys
// @Security   Bearer
// @Tags       App
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListApiKeyRequest	 true  "Param"
// @Success	   200		{object}	v1.ListApiKeyResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/api-keys [GET].
func (ctrl *ApiKeyHandler) List(c *gin.Context) {
	log.C(c).Infow("List apiKey function called")

	var req v1.ListApiKeyRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	resp, err := ctrl.b.ApiKeys().List(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Create
// @Summary    Create apiKey
// @Security   Bearer
// @Tags       App
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateApiKeyRequest	 true  "Param"
// @Success	   200		{object}	v1.ApiKeyInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/api-keys [POST].
func (ctrl *ApiKeyHandler) Create(c *gin.Context) {
	log.C(c).Infow("Create apiKey function called")

	var req v1.CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	// Create apiKey
	resp, err := ctrl.b.ApiKeys().Create(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Get
// @Summary    Get apiKey info
// @Security   Bearer
// @Tags       App
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.ApiKeyInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/api-keys/{id} [GET].
func (ctrl *ApiKeyHandler) Get(c *gin.Context) {
	log.C(c).Infow("Get apiKey function called")

	ID := cast.ToUint(c.Param("id"))
	apiKey, err := ctrl.b.ApiKeys().Get(c, ID)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, apiKey, nil)
}

// Update
// @Summary    Update apiKey info
// @Security   Bearer
// @Tags       App
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdateApiKeyRequest	 true  "Param"
// @Success	   200		{object}	v1.ApiKeyInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/api-keys/{id} [PUT].
func (ctrl *ApiKeyHandler) Update(c *gin.Context) {
	log.C(c).Infow("Update apiKey function called")

	var req v1.UpdateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.ApiKeys().Update(c, ID, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Delete
// @Summary    Delete apiKey
// @Security   Bearer
// @Tags       App
// @Accept     application/json
// @Produce    json
// @Param      id	    path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/api-keys/{id} [DELETE].
func (ctrl *ApiKeyHandler) Delete(c *gin.Context) {
	log.C(c).Infow("Delete apiKey function called")

	ID := cast.ToUint(c.Param("id"))
	if err := ctrl.b.ApiKeys().Delete(c, ID); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}
