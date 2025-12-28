// ABOUTME: HTTP handlers for TOTP security settings in AdminServer.
// ABOUTME: Provides endpoints for admin users to manage two-factor authentication.

package auth

import (
	"github.com/gin-gonic/gin"

	bizauth "github.com/bingo-project/bingo/internal/admserver/biz/auth"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/contextx"
)

// SecurityHandler handles TOTP security settings requests.
type SecurityHandler struct {
	securityBiz bizauth.SecurityBiz
}

// NewSecurityHandler creates a new SecurityHandler.
func NewSecurityHandler(ds store.IStore) *SecurityHandler {
	return &SecurityHandler{
		securityBiz: bizauth.NewSecurityBiz(ds),
	}
}

// GetTOTPStatus
// @Summary    Get TOTP status
// @Security   Bearer
// @Tags       Security
// @Accept     application/json
// @Produce    json
// @Success    200		{object}	v1.TOTPStatusResponse
// @Failure    400		{object}	core.ErrResponse
// @Failure    500		{object}	core.ErrResponse
// @Router     /v1/auth/security/totp/status [GET].
func (h *SecurityHandler) GetTOTPStatus(c *gin.Context) {
	log.C(c).Infow("GetTOTPStatus function called")

	username := contextx.Username(c)
	resp, err := h.securityBiz.GetTOTPStatus(c, username)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// SetupTOTP
// @Summary    Setup TOTP (get secret and QR code URL)
// @Security   Bearer
// @Tags       Security
// @Accept     application/json
// @Produce    json
// @Success    200		{object}	v1.TOTPSetupResponse
// @Failure    400		{object}	core.ErrResponse
// @Failure    500		{object}	core.ErrResponse
// @Router     /v1/auth/security/totp/setup [POST].
func (h *SecurityHandler) SetupTOTP(c *gin.Context) {
	log.C(c).Infow("SetupTOTP function called")

	username := contextx.Username(c)
	resp, err := h.securityBiz.SetupTOTP(c, username)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// EnableTOTP
// @Summary    Enable TOTP (verify code and enable)
// @Security   Bearer
// @Tags       Security
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.TOTPEnableRequest	 true  "Param"
// @Success    200		{object}	nil
// @Failure    400		{object}	core.ErrResponse
// @Failure    500		{object}	core.ErrResponse
// @Router     /v1/auth/security/totp/enable [POST].
func (h *SecurityHandler) EnableTOTP(c *gin.Context) {
	log.C(c).Infow("EnableTOTP function called")

	var req v1.TOTPEnableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	username := contextx.Username(c)
	if err := h.securityBiz.EnableTOTP(c, username, req.Code); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, gin.H{"message": "TOTP 启用成功"}, nil)
}

// VerifyTOTP
// @Summary    Verify TOTP code
// @Security   Bearer
// @Tags       Security
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.TOTPVerifyRequest	 true  "Param"
// @Success    200		{object}	nil
// @Failure    400		{object}	core.ErrResponse
// @Failure    500		{object}	core.ErrResponse
// @Router     /v1/auth/security/totp/verify [POST].
func (h *SecurityHandler) VerifyTOTP(c *gin.Context) {
	log.C(c).Infow("VerifyTOTP function called")

	var req v1.TOTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	username := contextx.Username(c)
	if err := h.securityBiz.VerifyTOTP(c, username, req.Code); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, gin.H{"valid": true}, nil)
}

// DisableTOTP
// @Summary    Disable TOTP
// @Security   Bearer
// @Tags       Security
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.TOTPDisableRequest	 true  "Param"
// @Success    200		{object}	nil
// @Failure    400		{object}	core.ErrResponse
// @Failure    500		{object}	core.ErrResponse
// @Router     /v1/auth/security/totp/disable [POST].
func (h *SecurityHandler) DisableTOTP(c *gin.Context) {
	log.C(c).Infow("DisableTOTP function called")

	var req v1.TOTPDisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	username := contextx.Username(c)
	if err := h.securityBiz.DisableTOTP(c, username, req.TOTPCode); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, gin.H{"message": "TOTP 已解绑"}, nil)
}
