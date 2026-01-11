// ABOUTME: HTTP handlers for AI Model management in admin panel.
// ABOUTME: Provides list/get/update endpoints for AI Model resources.
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

type ModelHandler struct {
	b biz.IBiz
}

func NewModelHandler(ds store.IStore) *ModelHandler {
	return &ModelHandler{b: biz.NewBiz(ds)}
}

// Create
// @Summary    Create AI model
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      request  body      v1.CreateAiModelRequest  true  "Create request"
// @Success    200      {object}  v1.AiModelInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/models [POST].
func (h *ModelHandler) Create(c *gin.Context) {
	var req v1.CreateAiModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))
		return
	}

	model, err := h.b.AiModels().Create(c, &req)
	core.Response(c, model, err)
}

// List
// @Summary    List AI models
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      providerName  query     string  false  "Filter by provider name"
// @Param      status        query     string  false  "Filter by status" Enums(active, disabled)
// @Success    200           {object}  v1.ListAiModelResponse
// @Failure    400           {object}  core.ErrResponse
// @Failure    500           {object}  core.ErrResponse
// @Router     /v1/ai/models [GET].
func (h *ModelHandler) List(c *gin.Context) {
	var req v1.ListAiModelRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	models, err := h.b.AiModels().List(c, &req)
	core.Response(c, models, err)
}

// Get
// @Summary    Get AI model by ID
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      id  path      int  true  "Model ID"
// @Success    200  {object}  v1.AiModelInfo
// @Failure    400  {object}  core.ErrResponse
// @Failure    404  {object}  core.ErrResponse
// @Router     /v1/ai/models/{id} [GET].
func (h *ModelHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("invalid model id"))

		return
	}

	model, err := h.b.AiModels().Get(c, uint(id))
	core.Response(c, model, err)
}

// Update
// @Summary    Update AI model
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      id        path      int                       true  "Model ID"
// @Param      request  body      v1.UpdateAiModelRequest   true  "Update request"
// @Success    200      {object}  v1.AiModelInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    404      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/models/{id} [PUT].
func (h *ModelHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("invalid model id"))

		return
	}

	var req v1.UpdateAiModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	model, err := h.b.AiModels().Update(c, uint(id), &req)
	core.Response(c, model, err)
}

// Delete
// @Summary    Delete AI model
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      id  path      int  true  "Model ID"
// @Success    200  {object}  nil
// @Failure    400  {object}  core.ErrResponse
// @Failure    404  {object}  core.ErrResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/ai/models/{id} [DELETE].
func (h *ModelHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("invalid model id"))
		return
	}

	err = h.b.AiModels().Delete(c, uint(id))
	core.Response(c, nil, err)
}
