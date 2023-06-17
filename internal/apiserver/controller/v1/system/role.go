package system

import (
	"github.com/asaskevich/govalidator"
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/biz"
	"bingo/internal/apiserver/global"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	v1 "bingo/pkg/api/bingo/v1"
	"bingo/pkg/auth"
)

// RoleController 是 role 模块在 Controller 层的实现，用来处理用户模块的请求.
type RoleController struct {
	a *auth.Authz
	b biz.IBiz
}

// NewRoleController 创建一个 role controller.
func NewRoleController(ds store.IStore, a *auth.Authz) *RoleController {
	return &RoleController{a: a, b: biz.NewBiz(ds)}
}

// List
//
// @Summary    List roles
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListRoleRequest	 true  "Param"
// @Success	   200		{object}	v1.ListRoleResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/roles [GET].
func (ctrl *RoleController) List(c *gin.Context) {
	log.C(c).Infow("List role function called")

	var r v1.ListRoleRequest
	if err := c.ShouldBindQuery(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	resp, err := ctrl.b.Roles().List(c, r.Offset, r.Limit)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Create
//
// @Summary    Create a role
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateRoleRequest	 true  "Param"
// @Success	   200		{object}	v1.GetRoleResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/roles [POST].
func (ctrl *RoleController) Create(c *gin.Context) {
	log.C(c).Infow("Create role function called")

	var r v1.CreateRoleRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	// Validator
	if _, err := govalidator.ValidateStruct(r); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create role
	resp, err := ctrl.b.Roles().Create(c, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Get
//
// @Summary    Get role info
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Role name"
// @Success	   200		{object}	v1.GetRoleResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/roles/{name} [GET].
func (ctrl *RoleController) Get(c *gin.Context) {
	log.C(c).Infow("Get role function called")

	roleName := c.Param("name")
	role, err := ctrl.b.Roles().Get(c, roleName)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, role)
}

// Update
//
// @Summary    Update role info
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string                  true  "Role name"
// @Param      request	 body	    v1.UpdateRoleRequest	true  "Param"
// @Success	   200		{object}	v1.GetRoleResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/roles/{name} [PUT].
func (ctrl *RoleController) Update(c *gin.Context) {
	log.C(c).Infow("Update role function called")

	var r v1.UpdateRoleRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(r); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	roleName := c.Param("name")
	resp, err := ctrl.b.Roles().Update(c, roleName, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete
//
// @Summary    Delete a role
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Role name"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/roles/{name} [DELETE].
func (ctrl *RoleController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete role function called")

	roleName := c.Param("name")
	if err := ctrl.b.Roles().Delete(c, roleName); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}

// SetPermissions
//
// @Summary    Set permissions
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Role name"
// @Param      request	 body	    v1.SetPermissionsRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/roles/{name}/permissions [PUT].
func (ctrl *RoleController) SetPermissions(c *gin.Context) {
	var r v1.SetPermissionsRequest
	if err := c.ShouldBind(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(r); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	name := c.Param("name")
	if name == global.RoleRoot {
		core.WriteResponse(c, errno.ErrForbidden, nil)

		return
	}

	err := ctrl.b.Roles().SetPermissions(c, ctrl.a, name, r.PermissionIDs)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}

// GetPermissionIDs
//
// @Summary    Get permissions
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      name      path      string           true  "Role name"
// @Success	   200		{object}	v1.GetPermissionIDsResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/roles/{name}/permissions [GET].
func (ctrl *RoleController) GetPermissionIDs(c *gin.Context) {
	name := c.Param("name")
	resp, err := ctrl.b.Roles().GetPermissionIDs(c, ctrl.a, name)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}
