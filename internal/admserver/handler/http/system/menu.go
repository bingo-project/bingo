package system

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"bingo/internal/admserver/biz"
	"bingo/internal/pkg/auth"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1"
)

type MenuHandler struct {
	a *auth.Authorizer
	b biz.IBiz
}

func NewMenuHandler(ds store.IStore, a *auth.Authorizer) *MenuHandler {
	return &MenuHandler{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List menus
// @Security   Bearer
// @Tags       Menu
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListMenuRequest	 true  "Param"
// @Success	   200		{object}	v1.ListMenuResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/menus [GET].
func (ctrl *MenuHandler) List(c *gin.Context) {
	log.C(c).Infow("List menu function called")

	var req v1.ListMenuRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument)

		return
	}

	resp, err := ctrl.b.Menus().List(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Create
// @Summary    Create menu
// @Security   Bearer
// @Tags       Menu
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateMenuRequest	 true  "Param"
// @Success	   200		{object}	v1.MenuInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/menus [POST].
func (ctrl *MenuHandler) Create(c *gin.Context) {
	log.C(c).Infow("Create menu function called")

	var req v1.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	// Create menu
	resp, err := ctrl.b.Menus().Create(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Get
// @Summary    Get menu info
// @Security   Bearer
// @Tags       Menu
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.MenuInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/menus/{id} [GET].
func (ctrl *MenuHandler) Get(c *gin.Context) {
	log.C(c).Infow("Get menu function called")

	ID := cast.ToUint(c.Param("id"))
	menu, err := ctrl.b.Menus().Get(c, ID)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, menu, nil)
}

// Update
// @Summary    Update menu info
// @Security   Bearer
// @Tags       Menu
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Param      request	 body	    v1.UpdateMenuRequest	 true  "Param"
// @Success	   200		{object}	v1.MenuInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/menus/{id} [PUT].
func (ctrl *MenuHandler) Update(c *gin.Context) {
	log.C(c).Infow("Update menu function called")

	var req v1.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.Menus().Update(c, ID, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Delete
// @Summary    Delete menu
// @Security   Bearer
// @Tags       Menu
// @Accept     application/json
// @Produce    json
// @Param      id	    path	    string            true  "ID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/menus/{id} [DELETE].
func (ctrl *MenuHandler) Delete(c *gin.Context) {
	log.C(c).Infow("Delete menu function called")

	ID := cast.ToUint(c.Param("id"))
	if err := ctrl.b.Menus().Delete(c, ID); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// Tree
// @Summary    Get menu tree
// @Security   Bearer
// @Tags       Menu
// @Accept     application/json
// @Produce    json
// @Success	   200		{object}	[]v1.MenuInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/menus/tree [GET].
func (ctrl *MenuHandler) Tree(c *gin.Context) {
	log.C(c).Infow("Tree menu function called")

	resp, err := ctrl.b.Menus().Tree(c)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// ToggleHidden
// @Summary    ToggleHidden
// @Security   Bearer
// @Tags       Menu
// @Accept     application/json
// @Produce    json
// @Param      id	     path	    string            		 true  "ID"
// @Success	   200		{object}	v1.MenuInfo
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/menus/{id}/toggle-hidden [POST].
func (ctrl *MenuHandler) ToggleHidden(c *gin.Context) {
	log.C(c).Infow("menu.ToggleHidden function called")

	ID := cast.ToUint(c.Param("id"))
	resp, err := ctrl.b.Menus().ToggleHidden(c, ID)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}
