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

type RoleController struct {
	a *auth.Authz
	b biz.IBiz
}

func NewRoleController(ds store.IStore, a *auth.Authz) *RoleController {
	return &RoleController{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List roles
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListRoleRequest	 true  "Param"
// @Success	   200		{object}	v1.ListResponse{data=[]v1.RoleInfo}
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/roles [GET].
func (ctrl *RoleController) List(c *gin.Context) {
	log.C(c).Infow("List role function called")

	var req v1.ListRoleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	resp, err := ctrl.b.Roles().List(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Create
// @Summary    Create a role
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateRoleRequest	 true  "Param"
// @Success	   200		{object}	v1.RoleInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/roles [POST].
func (ctrl *RoleController) Create(c *gin.Context) {
	log.C(c).Infow("Create role function called")

	var req v1.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	// Validator
	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create role
	resp, err := ctrl.b.Roles().Create(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Get
// @Summary    Get role info
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Role name"
// @Success	   200		{object}	v1.RoleInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/roles/{name} [GET].
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
// @Summary    Update role info
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string                  true  "Role name"
// @Param      request	 body	    v1.UpdateRoleRequest	true  "Param"
// @Success	   200		{object}	v1.RoleInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/roles/{name} [PUT].
func (ctrl *RoleController) Update(c *gin.Context) {
	log.C(c).Infow("Update role function called")

	var req v1.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	roleName := c.Param("name")
	resp, err := ctrl.b.Roles().Update(c, roleName, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete
// @Summary    Delete a role
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Role name"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/roles/{name} [DELETE].
func (ctrl *RoleController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete role function called")

	roleName := c.Param("name")
	if err := ctrl.b.Roles().Delete(c, roleName); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}

// SetApis
// @Summary    Set apis
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Role name"
// @Param      request	 body	    v1.SetApisRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/roles/{name}/apis [PUT].
func (ctrl *RoleController) SetApis(c *gin.Context) {
	var req v1.SetApisRequest
	if err := c.ShouldBind(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	name := c.Param("name")
	if name == global.RoleRoot {
		core.WriteResponse(c, errno.ErrForbidden, nil)

		return
	}

	err := ctrl.b.Roles().SetApis(c, ctrl.a, name, req.ApiIDs)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}

// GetApiIDs
// @Summary    Get apis
// @Security   Bearer
// @Tags       System.Role
// @Accept     application/json
// @Produce    json
// @Param      name      path      string           true  "Role name"
// @Success	   200		{object}	v1.GetApiIDsResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/roles/{name}/apis [GET].
func (ctrl *RoleController) GetApiIDs(c *gin.Context) {
	name := c.Param("name")
	resp, err := ctrl.b.Roles().GetApiIDs(c, ctrl.a, name)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}
