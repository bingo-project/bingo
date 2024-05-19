package auth

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	_ "bingo/internal/apiserver/http/request/v1"
	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
)

// Nonce
// @Summary	    Get Address auth nonce
// @Security	Bearer
// @Tags		Auth
// @Accept		application/json
// @Produce	    json
// @Param		request query       v1.AddressRequest    true	"ETH Address"
// @Success	    200		{object}	v1.NonceResponse
// @Failure	    400		{object}	core.ErrResponse
// @Failure	    500		{object}	core.ErrResponse
// @Router		/v1/auth/nonce [GET].
func (ctrl *AuthController) Nonce(c *gin.Context) {
	log.C(c).Infow("Nonce function called")

	var req v1.AddressRequest
	if err := c.ShouldBind(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	resp, err := ctrl.b.Auth().Nonce(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// LoginByAddress
// @Summary	    Login by provider
// @Security	Bearer
// @Tags		Auth
// @Accept		application/json
// @Produce	    json
// @Param		request     query	v1.LoginByAddressRequest	true	"Param"
// @Success	    200		{object}	v1.LoginResponse
// @Failure	    400		{object}	core.ErrResponse
// @Failure	    500		{object}	core.ErrResponse
// @Router		/v1/auth/login/address [POST].
func (ctrl *AuthController) LoginByAddress(c *gin.Context) {
	log.C(c).Infow("LoginByAddress function called")

	var req v1.LoginByAddressRequest
	if err := c.ShouldBind(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	resp, err := ctrl.b.Auth().LoginByAddress(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}
