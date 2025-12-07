package system

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

type AdminHandler struct {
	a *auth.Authz
	b biz.IBiz
}

func NewAdminHandler(ds store.IStore, a *auth.Authz) *AdminHandler {
	return &AdminHandler{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List admins
// @Security   Bearer
// @Tags       Admin
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListAdminRequest	 true  "Param"
// @Success	   200		{object}	v1.ListAdminResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/admins [GET].
func (ctrl *AdminHandler) List(c *gin.Context) {
	log.C(c).Infow("List admin function called")

	var req v1.ListAdminRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrBind)

		return
	}

	resp, err := ctrl.b.Admins().List(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Create
// @Summary    Create admin
// @Security   Bearer
// @Tags       Admin
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateAdminRequest	 true  "Param"
// @Success	   200		{object}	v1.AdminInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/admins [POST].
func (ctrl *AdminHandler) Create(c *gin.Context) {
	log.C(c).Infow("Create admin function called")

	var req v1.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	// Create admin
	resp, err := ctrl.b.Admins().Create(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Get
// @Summary    Get admin info
// @Security   Bearer
// @Tags       Admin
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Username"
// @Success	   200		{object}	v1.AdminInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/admins/{name} [GET].
func (ctrl *AdminHandler) Get(c *gin.Context) {
	log.C(c).Infow("Get admin function called")

	username := c.Param("name")
	admin, err := ctrl.b.Admins().Get(c, username)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, admin, nil)
}

// Update
// @Summary    Update admin info
// @Security   Bearer
// @Tags       Admin
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Username"
// @Param      request	 body	    v1.UpdateAdminRequest	 true  "Param"
// @Success	   200		{object}	v1.AdminInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/admins/{name} [PUT].
func (ctrl *AdminHandler) Update(c *gin.Context) {
	log.C(c).Infow("Update admin function called")

	var req v1.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	username := c.Param("name")
	resp, err := ctrl.b.Admins().Update(c, username, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Delete
// @Summary    Delete a admin
// @Security   Bearer
// @Tags       Admin
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string     true  "Username"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/admins/{name} [DELETE].
func (ctrl *AdminHandler) Delete(c *gin.Context) {
	log.C(c).Infow("Delete admin function called")

	username := c.Param("name")
	if err := ctrl.b.Admins().Delete(c, username); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// SetRoles
// @Summary    Set admin roles
// @Security   Bearer
// @Tags       Admin
// @Accept     application/json
// @Produce    json
// @Param      name      path       string  true  "Query params"
// @Param      request	 body	    v1.SetRolesRequest	 true  "Param"
// @Success	   200		{object}	v1.AdminInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/admins/{name}/roles [PUT].
func (ctrl *AdminHandler) SetRoles(c *gin.Context) {
	log.C(c).Infow("SetRoles function called")

	var req v1.SetRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	username := c.Param("name")
	resp, err := ctrl.b.Admins().SetRoles(c, username, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}
