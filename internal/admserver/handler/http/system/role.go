package system

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/admserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type RoleHandler struct {
	a *auth.Authorizer
	b biz.IBiz
}

func NewRoleHandler(ds store.IStore, a *auth.Authorizer) *RoleHandler {
	return &RoleHandler{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List roles
// @Security   Bearer
// @Tags       Role
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListRoleRequest	 true  "Param"
// @Success	   200		{object}	v1.ListRoleResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/roles [GET].
func (ctrl *RoleHandler) List(c *gin.Context) {
	log.C(c).Infow("List role function called")

	var req v1.ListRoleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument)

		return
	}

	resp, err := ctrl.b.Roles().List(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Create
// @Summary    Create a role
// @Security   Bearer
// @Tags       Role
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateRoleRequest	 true  "Param"
// @Success	   200		{object}	v1.RoleInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/roles [POST].
func (ctrl *RoleHandler) Create(c *gin.Context) {
	log.C(c).Infow("Create role function called")

	var req v1.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	// Create role
	resp, err := ctrl.b.Roles().Create(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Get
// @Summary    Get role info
// @Security   Bearer
// @Tags       Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Role name"
// @Success	   200		{object}	v1.RoleInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/roles/{name} [GET].
func (ctrl *RoleHandler) Get(c *gin.Context) {
	log.C(c).Infow("Get role function called")

	roleName := c.Param("name")
	role, err := ctrl.b.Roles().Get(c, roleName)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, role, nil)
}

// Update
// @Summary    Update role info
// @Security   Bearer
// @Tags       Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string                  true  "Role name"
// @Param      request	 body	    v1.UpdateRoleRequest	true  "Param"
// @Success	   200		{object}	v1.RoleInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/roles/{name} [PUT].
func (ctrl *RoleHandler) Update(c *gin.Context) {
	log.C(c).Infow("Update role function called")

	var req v1.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	roleName := c.Param("name")
	resp, err := ctrl.b.Roles().Update(c, roleName, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Delete
// @Summary    Delete a role
// @Security   Bearer
// @Tags       Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Role name"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/roles/{name} [DELETE].
func (ctrl *RoleHandler) Delete(c *gin.Context) {
	log.C(c).Infow("Delete role function called")

	roleName := c.Param("name")
	if err := ctrl.b.Roles().Delete(c, roleName); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// SetApis
// @Summary    Set apis
// @Security   Bearer
// @Tags       Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Role name"
// @Param      request	 body	    v1.SetApisRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/roles/{name}/apis [PUT].
func (ctrl *RoleHandler) SetApis(c *gin.Context) {
	var req v1.SetApisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	name := c.Param("name")
	err := ctrl.b.Roles().SetApis(c, ctrl.a, name, req.ApiIDs)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// GetApiIDs
// @Summary    Get apis
// @Security   Bearer
// @Tags       Role
// @Accept     application/json
// @Produce    json
// @Param      name      path      string           true  "Role name"
// @Success	   200		{object}	v1.GetApiIDsResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/roles/{name}/apis [GET].
func (ctrl *RoleHandler) GetApiIDs(c *gin.Context) {
	name := c.Param("name")
	resp, err := ctrl.b.Roles().GetApiIDs(c, ctrl.a, name)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// SetMenus
// @Summary    Set menus
// @Security   Bearer
// @Tags       Role
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Role name"
// @Param      request	 body	    v1.SetMenusRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/roles/{name}/menus [PUT].
func (ctrl *RoleHandler) SetMenus(c *gin.Context) {
	var req v1.SetMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	roleName := c.Param("name")
	err := ctrl.b.Roles().SetMenus(c, roleName, req.MenuIDs)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// GetMenuIDs
// @Summary    Get menuIDs of role
// @Security   Bearer
// @Tags       Role
// @Accept     application/json
// @Produce    json
// @Param      name      path      string           true  "Role name"
// @Success	   200		{object}	v1.GetApiIDsResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/roles/{name}/menus [GET].
func (ctrl *RoleHandler) GetMenuIDs(c *gin.Context) {
	roleName := c.Param("name")
	resp, err := ctrl.b.Roles().GetMenuIDs(c, roleName)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// All
// @Summary    All roles
// @Security   Bearer
// @Tags       Role
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	v1.ListRoleResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/roles/all [GET].
func (ctrl *RoleHandler) All(c *gin.Context) {
	log.C(c).Infow("All role function called")

	resp, err := ctrl.b.Roles().All(c)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}
