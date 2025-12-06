package system

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	v1 "bingo/pkg/api/apiserver/v1"
)

// Login returns a JWT token.
// @Summary	    Login
// @Security	Bearer
// @Tags		Auth
// @Accept		application/json
// @Produce	    json
// @Param		request	body		v1.LoginRequest	true	"Param"
// @Success	    200		{object}	v1.LoginResponse
// @Failure	    400		{object}	core.ErrResponse
// @Failure	    500		{object}	core.ErrResponse
// @Router		/v1/auth/login [POST].
func (ctrl *AdminController) Login(c *gin.Context) {
	log.C(c).Infow("Login function called")

	var req v1.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage("%s", err.Error()), nil)

		return
	}

	resp, err := ctrl.b.Admins().Login(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// ChangePassword
// @Summary    Change password
// @Security   Bearer
// @Tags       Admin
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string          	        true  "Username"
// @Param      request	 body	    v1.ChangePasswordRequest	true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/admins/{name}/change-password [PUT].
func (ctrl *AdminController) ChangePassword(c *gin.Context) {
	log.C(c).Infow("Change admin password function called")

	var req v1.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage("%s", err.Error()), nil)

		return
	}

	username := c.Param("name")
	err := ctrl.b.Admins().ChangePassword(c, username, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
