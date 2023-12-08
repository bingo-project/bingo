package system

import (
	"github.com/asaskevich/govalidator"
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/apiserver/biz"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	v1 "bingo/pkg/api/bingo/v1"
	"bingo/pkg/auth"
)

type PermissionController struct {
	a *auth.Authz
	b biz.IBiz
}

func NewPermissionController(ds store.IStore, a *auth.Authz) *PermissionController {
	return &PermissionController{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List permissions
// @Security   Bearer
// @Tags       System.Permission
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListPermissionRequest	 true  "Param"
// @Success	   200		{object}	v1.ListResponse{data=[]v1.PermissionInfo}
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/permissions [GET].
func (ctrl *PermissionController) List(c *gin.Context) {
	log.C(c).Infow("List permission function called")

	var req v1.ListPermissionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	resp, err := ctrl.b.Permissions().List(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Create
// @Summary    Create a permission
// @Security   Bearer
// @Tags       System.Permission
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreatePermissionRequest	 true  "Param"
// @Success	   200		{object}	v1.PermissionInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/permissions [POST].
func (ctrl *PermissionController) Create(c *gin.Context) {
	log.C(c).Infow("Create permission function called")

	var req v1.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	// Validator
	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create permission
	resp, err := ctrl.b.Permissions().Create(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Get
// @Summary    Get permission info
// @Security   Bearer
// @Tags       System.Permission
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.PermissionInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/permissions/{id} [GET].
func (ctrl *PermissionController) Get(c *gin.Context) {
	log.C(c).Infow("Get permission function called")

	ID := cast.ToUint(c.Param("id"))
	permission, err := ctrl.b.Permissions().Get(c, ID)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, permission)
}

// Update
// @Summary    Update permission info
// @Security   Bearer
// @Tags       System.Permission
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdatePermissionRequest	 true  "Param"
// @Success	   200		{object}	v1.PermissionInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/permissions/{id} [PUT].
func (ctrl *PermissionController) Update(c *gin.Context) {
	log.C(c).Infow("Update permission function called")

	var req v1.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.Permissions().Update(c, ID, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete
// @Summary    Delete a permission
// @Security   Bearer
// @Tags       System.Permission
// @Accept     application/json
// @Produce    json
// @Param      id	    path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/permissions/{id} [DELETE].
func (ctrl *PermissionController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete permission function called")

	ID := cast.ToUint(c.Param("id"))
	if err := ctrl.b.Permissions().Delete(c, ID); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}

// All
// @Summary    All permissions
// @Security   Bearer
// @Tags       System.Permission
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	[]v1.PermissionInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/permissions/all [GET].
func (ctrl *PermissionController) All(c *gin.Context) {
	log.C(c).Infow("All permission function called")

	resp, err := ctrl.b.Permissions().All(c)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}
