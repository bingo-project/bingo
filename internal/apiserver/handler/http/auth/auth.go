package auth

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/biz"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1"
	"bingo/pkg/auth"
	"bingo/pkg/contextx"
)

type AuthController struct {
	a *auth.Authz
	b biz.IBiz
}

func NewAuthController(ds store.IStore, a *auth.Authz) *AuthController {
	return &AuthController{a: a, b: biz.NewBiz(ds)}
}

// SendEmailCode
// @Summary    Send email code
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.SendEmailRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/code/email [POST].
func (ctrl *AuthController) SendEmailCode(c *gin.Context) {
	log.C(c).Infow("SendEmailCode function called")

	var req v1.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	err := ctrl.b.Email().SendEmailVerifyCode(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}

// Register
// @Summary    Register
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.RegisterRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/register [POST].
func (ctrl *AuthController) Register(c *gin.Context) {
	log.C(c).Infow("Register function called")

	var req v1.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	resp, err := ctrl.b.Auth().Register(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// UserInfo
// @Summary    Get user info
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	v1.UserInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/user-info [GET].
func (ctrl *AuthController) UserInfo(c *gin.Context) {
	log.C(c).Infow("UserInfo function called")

	user, ok := contextx.UserInfo[*v1.UserInfo](c.Request.Context())
	if !ok {
		core.WriteResponse(c, errno.ErrResourceNotFound, nil)

		return
	}

	core.WriteResponse(c, nil, user)
}
