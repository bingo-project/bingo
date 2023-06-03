package system

import (
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/apiserver/biz"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	v1 "bingo/pkg/api/bingo/v1"
	"bingo/pkg/auth"
)

// PermissionController 是 permission 模块在 Controller 层的实现，用来处理用户模块的请求.
type PermissionController struct {
	a *auth.Authz
	b biz.IBiz
}

// NewPermissionController 创建一个 permission controller.
func NewPermissionController(ds store.IStore, a *auth.Authz) *PermissionController {
	return &PermissionController{a: a, b: biz.NewBiz(ds)}
}

// List
//
// @Summary    List permissions
// @Security   Bearer
// @Tags       Permission
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListPermissionRequest	 true  "Param"
// @Success	   200		{object}	v1.ListPermissionResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/permissions [GET]
func (ctrl *PermissionController) List(c *gin.Context) {
	log.C(c).Infow("List permission function called")

	var r v1.ListPermissionRequest
	if err := c.ShouldBindQuery(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	resp, err := ctrl.b.Permissions().List(c, r.Offset, r.Limit)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Create
//
// @Summary    Create a permission
// @Security   Bearer
// @Tags       Permission
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreatePermissionRequest	 true  "Param"
// @Success	   200		{object}	v1.GetPermissionResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/permissions [POST]
func (ctrl *PermissionController) Create(c *gin.Context) {
	log.C(c).Infow("Create permission function called")

	var r v1.CreatePermissionRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	// Validator
	if _, err := govalidator.ValidateStruct(r); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create permission
	resp, err := ctrl.b.Permissions().Create(c, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Get
//
// @Summary    Get permission info
// @Security   Bearer
// @Tags       Permission
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.GetPermissionResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/permissions/{id} [GET]
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
//
// @Summary    Update permission info
// @Security   Bearer
// @Tags       Permission
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdatePermissionRequest	 true  "Param"
// @Success	   200		{object}	v1.GetPermissionResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/permissions/{id} [PUT]
func (ctrl *PermissionController) Update(c *gin.Context) {
	log.C(c).Infow("Update permission function called")

	var r v1.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(r); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.Permissions().Update(c, ID, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete
//
// @Summary    Delete a permission
// @Security   Bearer
// @Tags       Permission
// @Accept     application/json
// @Produce    json
// @Param      id	    path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/permissions/{id} [DELETE]
func (ctrl *PermissionController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete permission function called")

	ID := cast.ToUint(c.Param("id"))
	if err := ctrl.b.Permissions().Delete(c, ID); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
