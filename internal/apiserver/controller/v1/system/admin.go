package system

import (
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/biz"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	v1 "bingo/pkg/api/bingo/v1"
	"bingo/pkg/auth"
)

// AdminController 是 admin 模块在 Controller 层的实现，用来处理用户模块的请求.
type AdminController struct {
	a *auth.Authz
	b biz.IBiz
}

// NewAdminController 创建一个 admin controller.
func NewAdminController(ds store.IStore, a *auth.Authz) *AdminController {
	return &AdminController{a: a, b: biz.NewBiz(ds)}
}

// List
//
// @Summary    List admins
// @Security   Bearer
// @Tags       System.Admin
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListAdminRequest	 true  "Param"
// @Success	   200		{object}	v1.ListAdminResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/admins [GET]
func (ctrl *AdminController) List(c *gin.Context) {
	log.C(c).Infow("List admin function called")

	var r v1.ListAdminRequest
	if err := c.ShouldBindQuery(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	resp, err := ctrl.b.Admins().List(c, r.Offset, r.Limit)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Create
//
// @Summary    Create admin
// @Security   Bearer
// @Tags       System.Admin
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateAdminRequest	 true  "Param"
// @Success	   200		{object}	v1.GetAdminResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/admins [POST]
func (ctrl *AdminController) Create(c *gin.Context) {
	log.C(c).Infow("Create admin function called")

	var r v1.CreateAdminRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	// Validator
	if _, err := govalidator.ValidateStruct(r); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create admin
	resp, err := ctrl.b.Admins().Create(c, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Get
//
// @Summary    Get admin info
// @Security   Bearer
// @Tags       System.Admin
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Username"
// @Success	   200		{object}	v1.GetAdminResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/admins/{name} [GET]
func (ctrl *AdminController) Get(c *gin.Context) {
	log.C(c).Infow("Get admin function called")

	username := c.Param("name")
	admin, err := ctrl.b.Admins().Get(c, username)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, admin)
}

// Update
//
// @Summary    Update admin info
// @Security   Bearer
// @Tags       System.Admin
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Username"
// @Param      request	 body	    v1.UpdateAdminRequest	 true  "Param"
// @Success	   200		{object}	v1.GetAdminResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/admins/{name} [PUT]
func (ctrl *AdminController) Update(c *gin.Context) {
	log.C(c).Infow("Update admin function called")

	var r v1.UpdateAdminRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(r); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	username := c.Param("name")
	resp, err := ctrl.b.Admins().Update(c, username, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete
//
// @Summary    Delete a admin
// @Security   Bearer
// @Tags       System.Admin
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Username"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/admins/{name} [DELETE]
func (ctrl *AdminController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete admin function called")

	username := c.Param("name")
	if err := ctrl.b.Admins().Delete(c, username); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}

// SetRoles
//
// @Summary    Set admin roles
// @Security   Bearer
// @Tags       System.Admin
// @Accept     application/json
// @Produce    json
// @Param      name      path       string  true  "Query params"
// @Param      request	 body	    v1.SetRolesRequest	 true  "Param"
// @Success	   200		{object}	v1.GetAdminResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/admins/{name}/roles [PUT]
func (ctrl *AdminController) SetRoles(c *gin.Context) {
	log.C(c).Infow("SetRoles function called")

	var r v1.SetRolesRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(r); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	username := c.Param("name")
	resp, err := ctrl.b.Admins().SetRoles(c, username, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// SwitchRole
//
// @Summary    Switch role
// @Security   Bearer
// @Tags       System.Admin
// @Accept     application/json
// @Produce    json
// @Param      name      path       string  true  "Query params"
// @Param      request	 body	    v1.SwitchRoleRequest	 true  "Param"
// @Success	   200		{object}	v1.GetAdminResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/admins/{name}/switch-role [PUT]
func (ctrl *AdminController) SwitchRole(c *gin.Context) {
	log.C(c).Infow("SwitchRole function called")

	var r v1.SwitchRoleRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(r); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	username := c.Param("name")
	resp, err := ctrl.b.Admins().SwitchRole(c, username, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}
