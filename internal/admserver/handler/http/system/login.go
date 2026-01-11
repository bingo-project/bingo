package system

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
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
func (ctrl *AdminHandler) Login(c *gin.Context) {
	log.C(c).Infow("Login function called")

	var req v1.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	resp, err := ctrl.b.Admins().Login(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// LoginWithTOTP
// @Summary    Login with TOTP verification
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.TOTPLoginRequest	 true  "Param"
// @Success    200		{object}	v1.LoginResponse
// @Failure    400		{object}	core.ErrResponse
// @Failure    500		{object}	core.ErrResponse
// @Router     /v1/auth/login/totp [POST].
func (ctrl *AdminHandler) LoginWithTOTP(c *gin.Context) {
	log.C(c).Infow("LoginWithTOTP function called")

	var req v1.TOTPLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	resp, err := ctrl.b.Admins().LoginWithTOTP(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}
