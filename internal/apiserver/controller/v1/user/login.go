package user

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/component-base/log"

	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	v1 "bingo/pkg/api/bingo/v1"
)

// Login returns a JWT token.
//
// @Summary	    Login
// @Security	Bearer
// @Tags		Auth
// @Accept		application/json
// @Produce	    json
// @Param		request	body		v1.LoginRequest	true	"Param"
// @Success	    200		{object}	v1.LoginResponse
// @Failure	    400		{object}	core.ErrResponse
// @Failure	    500		{object}	core.ErrResponse
// @Router		/v1/login [POST].
func (ctrl *UserController) Login(c *gin.Context) {
	log.C(c).Infow("Login function called")

	var r v1.LoginRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	resp, err := ctrl.b.Users().Login(c, &r)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}
