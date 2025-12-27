package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/apiserver/biz"
	bizauth "github.com/bingo-project/bingo/internal/apiserver/biz/auth"
	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/contextx"
)

type AuthHandler struct {
	a                *auth.Authorizer
	b                biz.IBiz
	codeBiz          bizauth.CodeBiz
	resetPasswordBiz bizauth.ResetPasswordBiz
	userBiz          bizauth.UserBiz
	bindingsBiz      bizauth.BindingsBiz
}

func NewAuthHandler(ds store.IStore, a *auth.Authorizer) *AuthHandler {
	codeBiz := bizauth.NewCodeBiz(ds)
	return &AuthHandler{
		a:                a,
		b:                biz.NewBiz(ds),
		codeBiz:          codeBiz,
		resetPasswordBiz: bizauth.NewResetPasswordBiz(ds, codeBiz),
		userBiz:          bizauth.NewUserBiz(ds, codeBiz),
		bindingsBiz:      bizauth.NewBindingsBiz(ds),
	}
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

// SendCode 发送验证码
// @Summary    Send verification code
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.SendCodeRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/code [POST].
func (ctrl *AuthHandler) SendCode(c *gin.Context) {
	log.C(c).Infow("SendCode function called")

	var req v1.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	if err := ctrl.codeBiz.Send(c.Request.Context(), req.Account, bizauth.CodeScene(req.Scene)); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, gin.H{"message": "验证码已发送"}, nil)
}

// ResetPassword 重置密码
// @Summary    Reset password
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.ResetPasswordRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/reset-password [POST].
func (ctrl *AuthHandler) ResetPassword(c *gin.Context) {
	log.C(c).Infow("ResetPassword function called")

	var req v1.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	if err := ctrl.resetPasswordBiz.ResetPassword(c.Request.Context(), &req); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, gin.H{"message": "密码重置成功"}, nil)
}

// UpdateProfile 更新用户信息
// @Summary    Update user profile
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.UpdateProfileRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/user [PUT].
func (ctrl *AuthHandler) UpdateProfile(c *gin.Context) {
	log.C(c).Infow("UpdateProfile function called")

	var req v1.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	uid := contextx.UserID(c.Request.Context())
	if err := ctrl.userBiz.UpdateProfile(c.Request.Context(), uid, &req); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, gin.H{"message": "更新成功"}, nil)
}

// ListBindings 查询社交账号绑定
// @Summary    List social account bindings
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	v1.ListBindingsResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/bindings [GET].
func (ctrl *AuthHandler) ListBindings(c *gin.Context) {
	log.C(c).Infow("ListBindings function called")

	uid := contextx.UserID(c.Request.Context())
	resp, err := ctrl.bindingsBiz.ListBindings(c.Request.Context(), uid)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Unbind 解绑社交账号
// @Summary    Unbind social account
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      provider	 path	    string	 true  "Provider name"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/bindings/{provider} [DELETE].
func (ctrl *AuthHandler) Unbind(c *gin.Context) {
	log.C(c).Infow("Unbind function called")

	provider := c.Param("provider")
	uid := contextx.UserID(c.Request.Context())

	if err := ctrl.bindingsBiz.Unbind(c.Request.Context(), uid, provider); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, gin.H{"message": "解绑成功"}, nil)
}
