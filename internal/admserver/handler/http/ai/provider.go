// ABOUTME: HTTP handlers for AI Provider management in admin panel.
// ABOUTME: Provides list/get/update endpoints for AI Provider resources.
package ai

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/admserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type ProviderHandler struct {
	b biz.IBiz
}

func NewProviderHandler(ds store.IStore) *ProviderHandler {
	return &ProviderHandler{b: biz.NewBiz(ds)}
}

// List
// @Summary    List AI providers
// @Security   Bearer
// @Tags       AI Provider
// @Accept     application/json
// @Produce    json
// @Param      status   query     string  false  "Filter by status" Enums(active, disabled)
// @Success    200      {object}  v1.ListAiProviderResponse
// @Failure    400      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/providers [GET].
func (h *ProviderHandler) List(c *gin.Context) {
	var req v1.ListAiProviderRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	providers, err := h.b.AiProviders().List(c, &req)
	core.Response(c, providers, err)
}

// Get
// @Summary    Get AI provider by ID
// @Security   Bearer
// @Tags       AI Provider
// @Accept     application/json
// @Produce    json
// @Param      id  path      int  true  "Provider ID"
// @Success    200  {object}  v1.AiProviderInfo
// @Failure    400  {object}  core.ErrResponse
// @Failure    404  {object}  core.ErrResponse
// @Router     /v1/ai/providers/{id} [GET].
func (h *ProviderHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("invalid provider id"))

		return
	}

	provider, err := h.b.AiProviders().Get(c, uint(id))
	core.Response(c, provider, err)
}

// Update
// @Summary    Update AI provider
// @Security   Bearer
// @Tags       AI Provider
// @Accept     application/json
// @Produce    json
// @Param      id        path      int                           true  "Provider ID"
// @Param      request  body      v1.UpdateAiProviderRequest   true  "Update request"
// @Success    200      {object}  v1.AiProviderInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    404      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/providers/{id} [PUT].
func (h *ProviderHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("invalid provider id"))

		return
	}

	var req v1.UpdateAiProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	provider, err := h.b.AiProviders().Update(c, uint(id), &req)
	core.Response(c, provider, err)
}
