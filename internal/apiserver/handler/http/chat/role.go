// ABOUTME: HTTP handlers for AI role preset management.
// ABOUTME: Provides CRUD endpoints for AI role preset resources.

package chat

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/apiserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ai"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type RoleHandler struct {
	b biz.IBiz
}

func NewRoleHandler(ds store.IStore, registry *ai.Registry) *RoleHandler {
	return &RoleHandler{
		b: biz.NewBiz(ds).WithRegistry(registry),
	}
}

// Create
// @Summary    Create AI role
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      request  body      v1.CreateAiRoleRequest  true  "Param"
// @Success    200      {object}  v1.AiRoleInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/roles [POST].
func (h *RoleHandler) Create(c *gin.Context) {
	var req v1.CreateAiRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	role, err := h.b.AiRoles().Create(c, &req)
	core.Response(c, role, err)
}

// Get
// @Summary    Get AI role by role_id
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      role_id  path      string  true  "Role ID"
// @Success    200      {object}  v1.AiRoleInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    404      {object}  core.ErrResponse
// @Router     /v1/ai/roles/{role_id} [GET].
func (h *RoleHandler) Get(c *gin.Context) {
	roleID := c.Param("role_id")
	role, err := h.b.AiRoles().Get(c, roleID)
	core.Response(c, role, err)
}

// List
// @Summary    List AI roles
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      category  query     string  false  "Filter by category"
// @Param      status    query     string  false  "Filter by status"
// @Success    200       {object}  v1.ListAiRoleResponse
// @Failure    400       {object}  core.ErrResponse
// @Router     /v1/ai/roles [GET].
func (h *RoleHandler) List(c *gin.Context) {
	var req v1.ListAiRoleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	roles, err := h.b.AiRoles().List(c, &req)
	core.Response(c, roles, err)
}

// Update
// @Summary    Update AI role
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      role_id  path      string                   true  "Role ID"
// @Param      request  body      v1.UpdateAiRoleRequest   true  "Update request"
// @Success    200      {object}  v1.AiRoleInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    404      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/roles/{role_id} [PUT].
func (h *RoleHandler) Update(c *gin.Context) {
	var req v1.UpdateAiRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	roleID := c.Param("role_id")
	role, err := h.b.AiRoles().Update(c, roleID, &req)
	core.Response(c, role, err)
}

// Delete
// @Summary    Delete AI role
// @Security   Bearer
// @Tags       AI Role
// @Accept     application/json
// @Produce    json
// @Param      role_id  path  string  true  "Role ID"
// @Success    200      {object}  nil
// @Failure    404      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/roles/{role_id} [DELETE].
func (h *RoleHandler) Delete(c *gin.Context) {
	roleID := c.Param("role_id")
	err := h.b.AiRoles().Delete(c, roleID)
	core.Response(c, nil, err)
}
