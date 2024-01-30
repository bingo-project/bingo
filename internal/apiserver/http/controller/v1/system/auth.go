package system

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/apiserver/biz"
	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/pkg/auth"
)

type AuthController struct {
	a *auth.Authz
	b biz.IBiz
}

func NewAuthController(ds store.IStore, a *auth.Authz) *AuthController {
	return &AuthController{a: a, b: biz.NewBiz(ds)}
}

// UserInfo
// @Summary    Get user info
// @Security   Bearer
// @Tags       System.Auth
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	v1.AdminInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/auth/user-info [GET].
func (ctrl *AuthController) UserInfo(c *gin.Context) {
	log.C(c).Infow("Self function called")

	var admin v1.AdminInfo
	err := auth.User(c, &admin)
	if err != nil {
		core.WriteResponse(c, errno.ErrResourceNotFound, nil)

		return
	}

	core.WriteResponse(c, nil, admin)
}

// Menus
// @Summary    Get menu tree
// @Security   Bearer
// @Tags       System.Auth
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	v1.ListMenuResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/auth/menus [GET].
func (ctrl *AuthController) Menus(c *gin.Context) {
	log.C(c).Infow("Menus function called")

	var admin v1.AdminInfo
	_ = auth.User(c, &admin)

	resp, err := ctrl.b.Roles().GetMenuTree(c, admin.RoleName)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// ChangePassword
// @Summary    Change password
// @Security   Bearer
// @Tags       System.Auth
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.ChangePasswordRequest	true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/auth/change-password [PUT].
func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	log.C(c).Infow("Change admin password function called")

	var req v1.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	username := cast.ToString(auth.ID(c))
	err := ctrl.b.Admins().ChangePassword(c, username, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}

// SwitchRole
// @Summary    Switch role
// @Security   Bearer
// @Tags       System.Auth
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.SwitchRoleRequest	 true  "Param"
// @Success	   200		{object}	v1.AdminInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/auth/switch-role [PUT].
func (ctrl *AuthController) SwitchRole(c *gin.Context) {
	log.C(c).Infow("SwitchRole function called")

	var req v1.SwitchRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	username := cast.ToString(auth.ID(c))
	resp, err := ctrl.b.Admins().SwitchRole(c, username, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}
