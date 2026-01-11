package user

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/admserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type UserHandler struct {
	a *auth.Authorizer
	b biz.IBiz
}

func NewUserHandler(ds store.IStore, a *auth.Authorizer) *UserHandler {
	return &UserHandler{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List users
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListUserRequest	 true  "Param"
// @Param      keyword	 query	    string	 false  "Search keyword for UID/Username/Email/Phone"
// @Success	   200		{object}	v1.ListUserResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/users [GET].
func (ctrl *UserHandler) List(c *gin.Context) {
	log.C(c).Infow("List user function called")

	var req v1.ListUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument)

		return
	}

	resp, err := ctrl.b.Users().List(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Create
// @Summary    Create a user
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateUserRequest	 true  "Param"
// @Success	   200		{object}	v1.UserInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/users [POST].
func (ctrl *UserHandler) Create(c *gin.Context) {
	log.C(c).Infow("Create user function called")

	var req v1.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	// Create user
	resp, err := ctrl.b.Users().Create(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	// Create policy
	if _, err := ctrl.a.Enforcer().AddNamedPolicy("p", req.Username, "/v1/users/"+req.Username, auth.AclDefaultMethods); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Get
// @Summary    Get user info
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      uid	     path	    string          	 true  "User UID"
// @Success	   200		{object}	v1.UserInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/users/{uid} [GET].
func (ctrl *UserHandler) Get(c *gin.Context) {
	log.C(c).Infow("Get user function called")

	user, err := ctrl.b.Users().Get(c, c.Param("uid"))
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, user, nil)
}

// Update
// @Summary    Update user info
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      uid	     path	    string          	 true  "User UID"
// @Param      request	 body	    v1.UpdateUserRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/users/{uid} [PUT].
func (ctrl *UserHandler) Update(c *gin.Context) {
	log.C(c).Infow("Update user function called")

	var req v1.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument)

		return
	}

	if err := ctrl.b.Users().Update(c, c.Param("uid"), &req); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// Delete
// @Summary    Delete a user
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      uid	     path	    string          	 true  "User UID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/users/{uid} [DELETE].
func (ctrl *UserHandler) Delete(c *gin.Context) {
	log.C(c).Infow("Delete user function called")

	uid := c.Param("uid")

	// Get user info first (need username for ACL cleanup)
	user, err := ctrl.b.Users().Get(c, uid)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	// Delete user
	if err := ctrl.b.Users().Delete(c, uid); err != nil {
		core.Response(c, nil, err)

		return
	}

	// Remove ACL policy
	if _, err := ctrl.a.Enforcer().RemoveFilteredNamedPolicy("p", 0, user.Username); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// ResetPassword
// @Summary    Reset user password
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      uid	     path	    string                    true  "User UID"
// @Param      request	 body	    v1.ResetUserPasswordRequest true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/users/{uid}/password [PUT].
func (ctrl *UserHandler) ResetPassword(c *gin.Context) {
	log.C(c).Infow("ResetPassword function called")

	uid := c.Param("uid")
	var req v1.ResetUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	if err := ctrl.b.Users().ResetPassword(c, uid, req.Password); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}
