// ABOUTME: HTTP handlers for security settings (pay password and TOTP).
// ABOUTME: Provides endpoints for sensitive operation verification.

package auth

import (
	"github.com/gin-gonic/gin"

	bizauth "github.com/bingo-project/bingo/internal/apiserver/biz/auth"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/contextx"
)

// SecurityHandler handles security settings requests.
type SecurityHandler struct {
	securityBiz bizauth.SecurityBiz
}

// NewSecurityHandler creates a new SecurityHandler.
func NewSecurityHandler(ds store.IStore) *SecurityHandler {
	codeBiz := bizauth.NewCodeBiz(ds)
	return &SecurityHandler{
		securityBiz: bizauth.NewSecurityBiz(ds, codeBiz),
	}
}

// GetSecurityStatus
// @Summary    Get security status
// @Security   Bearer
// @Tags       Security
// @Accept     application/json
// @Produce    json
// @Success    200		{object}	v1.SecurityStatusResponse
// @Failure    400		{object}	core.ErrResponse
// @Failure    500		{object}	core.ErrResponse
// @Router     /v1/auth/security/status [GET].
func (h *SecurityHandler) GetSecurityStatus(c *gin.Context) {
	log.C(c).Infow("GetSecurityStatus function called")

	uid := contextx.UserID(c.Request.Context())
	resp, err := h.securityBiz.GetSecurityStatus(c.Request.Context(), uid)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// SetPayPassword
// @Summary    Set pay password
// @Security   Bearer
// @Tags       Security
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.SetPayPasswordRequest	 true  "Param"
// @Success    200		{object}	nil
// @Failure    400		{object}	core.ErrResponse
// @Failure    500		{object}	core.ErrResponse
// @Router     /v1/auth/security/pay-password [PUT].
func (h *SecurityHandler) SetPayPassword(c *gin.Context) {
	log.C(c).Infow("SetPayPassword function called")

	var req v1.SetPayPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	uid := contextx.UserID(c.Request.Context())
	if err := h.securityBiz.SetPayPassword(c.Request.Context(), uid, &req); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, gin.H{"message": "支付密码设置成功"}, nil)
}

// VerifyPayPassword
// @Summary    Verify pay password
// @Security   Bearer
// @Tags       Security
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.VerifyPayPasswordRequest	 true  "Param"
// @Success    200		{object}	nil
// @Failure    400		{object}	core.ErrResponse
// @Failure    500		{object}	core.ErrResponse
// @Router     /v1/auth/security/verify-pay-password [POST].
func (h *SecurityHandler) VerifyPayPassword(c *gin.Context) {
	log.C(c).Infow("VerifyPayPassword function called")

	var req v1.VerifyPayPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	uid := contextx.UserID(c.Request.Context())
	if err := h.securityBiz.VerifyPayPassword(c.Request.Context(), uid, req.PayPassword); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, gin.H{"valid": true}, nil)
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

	uid := contextx.UserID(c.Request.Context())
	resp, err := h.securityBiz.GetTOTPStatus(c.Request.Context(), uid)
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

	uid := contextx.UserID(c.Request.Context())
	user, ok := contextx.UserInfo[*v1.UserInfo](c.Request.Context())
	email := ""
	if ok && user != nil {
		email = user.Email
	}

	resp, err := h.securityBiz.SetupTOTP(c.Request.Context(), uid, email)
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

	uid := contextx.UserID(c.Request.Context())
	if err := h.securityBiz.EnableTOTP(c.Request.Context(), uid, req.Code); err != nil {
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

	uid := contextx.UserID(c.Request.Context())
	if err := h.securityBiz.VerifyTOTP(c.Request.Context(), uid, req.Code); err != nil {
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

	uid := contextx.UserID(c.Request.Context())
	if err := h.securityBiz.DisableTOTP(c.Request.Context(), uid, req.VerifyCode, req.TOTPCode); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, gin.H{"message": "TOTP 已解绑"}, nil)
}
