package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/component-base/log"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/model"
	"bingo/pkg/api/apiserver/v1"
	"bingo/pkg/auth"
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
func (ctrl *AuthController) Login(c *gin.Context) {
	log.C(c).Infow("Login function called")

	var req v1.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	resp, err := ctrl.b.Auth().Login(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// GetAuthCode
// @Summary	    Login by provider
// @Security	Bearer
// @Tags		Auth
// @Accept		application/json
// @Produce	    json
// @Param		provider    path    string          true	"Auth provider name"
// @Param		request     query	v1.LoginByProviderRequest	true	"Param"
// @Success	    200		{object}	v1.LoginResponse
// @Failure	    400		{object}	core.ErrResponse
// @Failure	    500		{object}	core.ErrResponse
// @Router		/v1/auth/login/{provider} [GET].
func (ctrl *AuthController) GetAuthCode(c *gin.Context) {
	log.C(c).Infow("LoginByProvider function called")

	var req v1.LoginByProviderRequest
	if err := c.ShouldBind(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	core.WriteResponse(c, nil, req)
}

// LoginByProvider
// @Summary	    Login by provider
// @Security	Bearer
// @Tags		Auth
// @Accept		application/json
// @Produce	    json
// @Param		provider    path    string          true	"Auth provider name"
// @Param		request     query	v1.LoginByProviderRequest	true	"Param"
// @Success	    200		{object}	v1.LoginResponse
// @Failure	    400		{object}	core.ErrResponse
// @Failure	    500		{object}	core.ErrResponse
// @Router		/v1/auth/login/{provider} [POST].
func (ctrl *AuthController) LoginByProvider(c *gin.Context) {
	log.C(c).Infow("LoginByProvider function called")

	var req v1.LoginByProviderRequest
	if err := c.ShouldBind(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	provider := c.Param("provider")
	resp, err := ctrl.b.Auth().LoginByProvider(c, provider, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// BindProvider
// @Summary	    Bind provider
// @Security	Bearer
// @Tags		Auth
// @Accept		application/json
// @Produce	    json
// @Param		provider    path    string          true	"Auth provider name"
// @Param		request     query	v1.LoginByProviderRequest	true	"Param"
// @Success	    200		{object}	v1.LoginResponse
// @Failure	    400		{object}	core.ErrResponse
// @Failure	    500		{object}	core.ErrResponse
// @Router		/v1/auth/bind/{provider} [POST].
func (ctrl *AuthController) BindProvider(c *gin.Context) {
	log.C(c).Infow("BindProvider function called")

	var req v1.LoginByProviderRequest
	if err := c.ShouldBind(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	var user model.UserM
	_ = auth.User(c, &user)

	provider := c.Param("provider")
	resp, err := ctrl.b.Auth().Bind(c, provider, &req, &user)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}
