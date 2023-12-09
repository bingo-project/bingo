package system

import (
	"github.com/asaskevich/govalidator"
	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/apiserver/biz"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	v1 "bingo/pkg/api/bingo/v1"
	"bingo/pkg/auth"
)

type MenuController struct {
	a *auth.Authz
	b biz.IBiz
}

func NewMenuController(ds store.IStore, a *auth.Authz) *MenuController {
	return &MenuController{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List menus
// @Security   Bearer
// @Tags       System.Menu
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListMenuRequest	 true  "Param"
// @Success	   200		{object}	v1.ListResponse{data=[]v1.MenuInfo}
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/menus [GET]
func (ctrl *MenuController) List(c *gin.Context) {
	log.C(c).Infow("List menu function called")

	var req v1.ListMenuRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	resp, err := ctrl.b.Menus().List(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Create
// @Summary    Create menu
// @Security   Bearer
// @Tags       System.Menu
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateMenuRequest	 true  "Param"
// @Success	   200		{object}	v1.MenuInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/menus [POST]
func (ctrl *MenuController) Create(c *gin.Context) {
	log.C(c).Infow("Create menu function called")

	var req v1.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	// Validator
	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	// Create menu
	resp, err := ctrl.b.Menus().Create(c, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Get
// @Summary    Get menu info
// @Security   Bearer
// @Tags       System.Menu
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.MenuInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/menus/{id} [GET]
func (ctrl *MenuController) Get(c *gin.Context) {
	log.C(c).Infow("Get menu function called")

	ID := cast.ToUint(c.Param("id"))
	menu, err := ctrl.b.Menus().Get(c, ID)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, menu)
}

// Update
// @Summary    Update menu info
// @Security   Bearer
// @Tags       System.Menu
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdateMenuRequest	 true  "Param"
// @Success	   200		{object}	v1.MenuInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/menus/{id} [PUT]
func (ctrl *MenuController) Update(c *gin.Context) {
	log.C(c).Infow("Update menu function called")

	var req v1.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)

		return
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(err.Error()), nil)

		return
	}

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.Menus().Update(c, ID, &req)
	if err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, resp)
}

// Delete
// @Summary    Delete menu
// @Security   Bearer
// @Tags       System.Menu
// @Accept     application/json
// @Produce    json
// @Param      id	    path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/system/menus/{id} [DELETE]
func (ctrl *MenuController) Delete(c *gin.Context) {
	log.C(c).Infow("Delete menu function called")

	ID := cast.ToUint(c.Param("id"))
	if err := ctrl.b.Menus().Delete(c, ID); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
