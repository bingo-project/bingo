// ABOUTME: HTTP handlers for AI role preset management.
// ABOUTME: Provides read-only endpoints for AI role preset resources.

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
