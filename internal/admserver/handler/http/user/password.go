package user

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

// ChangePassword 修改指定用户的密码.
// @Summary    Change password
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string          	        true  "Username"
// @Param      request	 body	    v1.ChangePasswordRequest	true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/users/{name}/change-password [PUT].
func (ctrl *UserHandler) ChangePassword(c *gin.Context) {
	log.C(c).Infow("Change password function called")

	var req v1.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	username := c.Param("name")
	err := ctrl.b.Auth().ChangePassword(c, username, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}
