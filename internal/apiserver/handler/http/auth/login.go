package auth

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	v1 "bingo/pkg/api/apiserver/v1"
	"bingo/pkg/contextx"
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
func (ctrl *AuthHandler) Login(c *gin.Context) {
	log.C(c).Infow("Login function called")

	var req v1.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	resp, err := ctrl.b.Auth().Login(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
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
func (ctrl *AuthHandler) GetAuthCode(c *gin.Context) {
	log.C(c).Infow("LoginByProvider function called")

	var req v1.LoginByProviderRequest
	if err := c.ShouldBind(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	core.Response(c, req, nil)
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
func (ctrl *AuthHandler) LoginByProvider(c *gin.Context) {
	log.C(c).Infow("LoginByProvider function called")

	var req v1.LoginByProviderRequest
	if err := c.ShouldBind(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	provider := c.Param("provider")
	resp, err := ctrl.b.Auth().LoginByProvider(c, provider, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
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
func (ctrl *AuthHandler) BindProvider(c *gin.Context) {
	log.C(c).Infow("BindProvider function called")

	var req v1.LoginByProviderRequest
	if err := c.ShouldBind(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	user, _ := contextx.UserInfo[*v1.UserInfo](c.Request.Context())

	provider := c.Param("provider")
	resp, err := ctrl.b.Auth().Bind(c, provider, &req, user)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}
