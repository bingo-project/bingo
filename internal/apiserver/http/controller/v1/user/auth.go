package user

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/apiserver/biz"
	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/pkg/auth"
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
// @Router    /v1/auth/code/email [POST]
func (ctrl *AuthController) SendEmailCode(c *gin.Context) {
	log.C(c).Infow("Login function called")

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
