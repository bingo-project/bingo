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
	"github.com/bingo-project/bingo/pkg/contextx"
)

type AuthHandler struct {
	a *auth.Authorizer
	b biz.IBiz
}

func NewAuthHandler(ds store.IStore, a *auth.Authorizer) *AuthHandler {
	return &AuthHandler{a: a, b: biz.NewBiz(ds)}
}

// UserInfo
// @Summary    Get user info
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	v1.AdminInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/user-info [GET].
func (ctrl *AuthHandler) UserInfo(c *gin.Context) {
	log.C(c).Infow("UserInfo function called")

	admin, ok := contextx.UserInfo[*v1.AdminInfo](c.Request.Context())
	if !ok {
		core.Response(c, nil, errno.ErrNotFound)

		return
	}

	core.Response(c, *admin, nil)
}

// Menus
// @Summary    Get menu tree
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	v1.ListMenuResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/menus [GET].
func (ctrl *AuthHandler) Menus(c *gin.Context) {
	log.C(c).Infow("Menus function called")

	admin, _ := contextx.UserInfo[v1.AdminInfo](c.Request.Context())

	resp, err := ctrl.b.Roles().GetMenuTree(c, admin.RoleName)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// ChangePassword
// @Summary    Change password
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.ChangePasswordRequest	true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/change-password [PUT].
func (ctrl *AuthHandler) ChangePassword(c *gin.Context) {
	log.C(c).Infow("Change admin password function called")

	var req v1.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	username := contextx.Username(c.Request.Context())
	err := ctrl.b.Admins().ChangePassword(c, username, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// SwitchRole
// @Summary    Switch role
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.SwitchRoleRequest	 true  "Param"
// @Success	   200		{object}	v1.AdminInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/switch-role [PUT].
func (ctrl *AuthHandler) SwitchRole(c *gin.Context) {
	log.C(c).Infow("SwitchRole function called")

	var req v1.SwitchRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	username := contextx.Username(c.Request.Context())
	resp, err := ctrl.b.Admins().SwitchRole(c, username, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}
