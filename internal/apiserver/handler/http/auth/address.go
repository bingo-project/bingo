// ABOUTME: HTTP handlers for wallet authentication.
// ABOUTME: Provides SIWE nonce generation and wallet login endpoints.

package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

// Nonce
// @Summary     Get SIWE nonce for wallet login
// @Tags        Auth
// @Accept      application/json
// @Produce     json
// @Param       request query       v1.AddressRequest    true  "ETH Address"
// @Success     200     {object}    v1.NonceResponse
// @Failure     400     {object}    core.ErrResponse
// @Failure     500     {object}    core.ErrResponse
// @Router      /v1/auth/nonce [GET].
func (h *AuthHandler) Nonce(c *gin.Context) {
	var req v1.AddressRequest
	if err := c.ShouldBind(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
		return
	}

	resp, err := h.b.Auth().Nonce(c, &req)
	core.Response(c, resp, err)
}

// LoginByAddress
// @Summary     Wallet login with SIWE
// @Tags        Auth
// @Accept      application/json
// @Produce     json
// @Param       request  body      v1.LoginByAddressRequest  true  "Param"
// @Success     200      {object}  v1.LoginResponse
// @Failure     400      {object}  core.ErrResponse
// @Failure     401      {object}  core.ErrResponse
// @Router      /v1/auth/login/address [POST].
func (h *AuthHandler) LoginByAddress(c *gin.Context) {
	var req v1.LoginByAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
		return
	}

	resp, err := h.b.Auth().LoginByAddress(c, &req)
	core.Response(c, resp, err)
}
