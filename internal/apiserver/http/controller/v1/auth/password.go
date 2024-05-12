package auth

import (
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/pkg/auth"
)

// ChangePassword 修改指定用户的密码.
// @Summary    Change password
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.ChangePasswordRequest	true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/auth/change-password [PUT].
func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	log.C(c).Infow("ChangePassword function called")

	var req v1.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	uid := auth.ID(c)
	err := ctrl.b.Auth().ChangePassword(c, cast.ToString(uid), &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
