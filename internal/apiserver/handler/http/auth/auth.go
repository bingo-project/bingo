package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/apiserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/contextx"
)

type AuthHandler struct {
	a *auth.Authorizer
	b biz.IBiz
}

func NewAuthHandler(ds store.IStore, a *auth.Authorizer) *AuthHandler {
	return &AuthHandler{a: a, b: biz.NewBiz(ds)}
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
func (ctrl *AuthHandler) SendEmailCode(c *gin.Context) {
	log.C(c).Infow("SendEmailCode function called")

	var req v1.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	err := ctrl.b.Email().SendEmailVerifyCode(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
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
func (ctrl *AuthHandler) Register(c *gin.Context) {
	log.C(c).Infow("Register function called")

	var req v1.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	resp, err := ctrl.b.Auth().Register(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
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
func (ctrl *AuthHandler) UserInfo(c *gin.Context) {
	log.C(c).Infow("UserInfo function called")

	user, ok := contextx.UserInfo[*v1.UserInfo](c.Request.Context())
	if !ok {
		core.Response(c, nil, errno.ErrNotFound)

		return
	}

	core.Response(c, user, nil)
}
