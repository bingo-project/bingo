package system

import (
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	v1 "bingo/pkg/api/bingo/v1"
)

// Login returns a JWT token.
//
// @Summary	    Login
// @Security	Bearer
// @Tags		System.Auth
// @Accept		application/json
// @Produce	    json
// @Param		request	body		v1.LoginRequest	true	"Param"
// @Success	    200		{object}	v1.LoginResponse
// @Failure	    400		{object}	core.ErrResponse
// @Failure	    500		{object}	core.ErrResponse
// @Router		/system/login [POST].
func (ctrl *AdminController) Login(c *gin.Context) {
	log.C(c).Infow("Login function called")

	var r v1.LoginRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	resp, err := ctrl.b.Admins().Login(c, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// ChangePassword
//
// @Summary    Change password
// @Security   Bearer
// @Tags       System.Admin
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string          	        true  "Username"
// @Param      request	 body	    v1.ChangePasswordRequest	true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /system/admin/{name}/change-password [PUT].
func (ctrl *AdminController) ChangePassword(c *gin.Context) {
	log.C(c).Infow("Change admin password function called")

	var r v1.ChangePasswordRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(r); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	username := c.Param("name")
	err := ctrl.b.Users().ChangePassword(c, username, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
