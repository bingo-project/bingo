package user

import (
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/component-base/log"

	"bingo/internal/apiserver/biz"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	v1 "bingo/pkg/api/bingo/v1"
	"bingo/pkg/auth"
)

const defaultMethods = "(GET)|(POST)|(PUT)|(DELETE)"

// UserController 是 user 模块在 Controller 层的实现，用来处理用户模块的请求.
type UserController struct {
	a *auth.Authz
	b biz.IBiz
}

// NewUserController 创建一个 user controller.
func NewUserController(ds store.IStore, a *auth.Authz) *UserController {
	return &UserController{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List users
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListUserRequest	 true  "Param"
// @Success	   200		{object}	v1.ListResponse{data=[]v1.UserInfo}
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/users [GET].
func (ctrl *UserController) List(c *gin.Context) {
	log.C(c).Infow("List user function called")

	var req v1.ListUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	resp, err := ctrl.b.Users().List(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
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
func (ctrl *UserController) Create(c *gin.Context) {
	log.C(c).Infow("Create user function called")

	var req v1.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	// Validator
	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create user
	if err := ctrl.b.Users().Create(c, &req); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	// Create policy
	if _, err := ctrl.a.AddNamedPolicy("p", req.Username, "/v1/users/"+req.Username, defaultMethods); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}

// Get
// @Summary    Get user info
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string          	 true  "Username"
// @Success	   200		{object}	v1.UserInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/users/{name} [GET].
func (ctrl *UserController) Get(c *gin.Context) {
	log.C(c).Infow("Get user function called")

	user, err := ctrl.b.Users().Get(c, c.Param("name"))
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, user)
}

// Update
// @Summary    Update user info
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string          	 true  "Username"
// @Param      request	 query	    v1.UpdateUserRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/users/{name} [PUT].
func (ctrl *UserController) Update(c *gin.Context) {
	log.C(c).Infow("Update user function called")

	var req v1.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	if err := ctrl.b.Users().Update(c, c.Param("name"), &req); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}

// Delete
// @Summary    Delete a user
// @Security   Bearer
// @Tags       User
// @Accept     application/json
// @Produce    json
// @Param      name	     path	    string          	 true  "Username"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/users/{name} [DELETE].
func (ctrl *UserController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete user function called")

	username := c.Param("name")

	if err := ctrl.b.Users().Delete(c, username); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	if _, err := ctrl.a.RemoveNamedPolicy("p", username, "", ""); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
