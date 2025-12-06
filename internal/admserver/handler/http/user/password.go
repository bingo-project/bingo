package user

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	v1 "bingo/pkg/api/apiserver/v1"
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
func (ctrl *UserController) ChangePassword(c *gin.Context) {
	log.C(c).Infow("Change password function called")

	var req v1.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage("%s", err.Error()), nil)

		return
	}

	username := c.Param("name")
	err := ctrl.b.Auth().ChangePassword(c, username, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
