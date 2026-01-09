// ABOUTME: HTTP handlers for AI User Quota management in admin panel.
// ABOUTME: Provides list/get/update/reset endpoints for user quota resources.
package ai

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/admserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type QuotaHandler struct {
	b biz.IBiz
}

func NewQuotaHandler(ds store.IStore) *QuotaHandler {
	return &QuotaHandler{b: biz.NewBiz(ds)}
}

// List
// @Summary    List user quotas
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      tier      query     string  false  "Filter by tier" Enums(free, pro, enterprise)
// @Param      uid       query     string  false  "Filter by UID (prefix match)"
// @Param      page      query     int     false  "Page number" minimum(1)
// @Param      pageSize  query     int     false  "Page size" minimum(1) maximum(100)
// @Success    200       {object}  v1.ListAiUserQuotaResponse
// @Failure    400       {object}  core.ErrResponse
// @Failure    500       {object}  core.ErrResponse
// @Router     /v1/ai/quotas [GET].
func (h *QuotaHandler) List(c *gin.Context) {
	var req v1.ListAiUserQuotaRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	quotas, err := h.b.AiQuotas().List(c, &req)
	core.Response(c, quotas, err)
}

// Get
// @Summary    Get user quota by UID
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      uid  path      string  true  "User UID"
// @Success    200  {object}  v1.AiUserQuotaInfo
// @Failure    400  {object}  core.ErrResponse
// @Failure    404  {object}  core.ErrResponse
// @Router     /v1/ai/quotas/{uid} [GET].
func (h *QuotaHandler) Get(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("uid is required"))

		return
	}

	quota, err := h.b.AiQuotas().Get(c, uid)
	core.Response(c, quota, err)
}

// Update
// @Summary    Update user quota
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      uid       path      string                       true  "User UID"
// @Param      request  body      v1.UpdateAiUserQuotaRequest  true  "Update request"
// @Success    200      {object}  v1.AiUserQuotaInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    404      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/quotas/{uid} [PUT].
func (h *QuotaHandler) Update(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("uid is required"))

		return
	}

	var req v1.UpdateAiUserQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	quota, err := h.b.AiQuotas().Update(c, uid, &req)
	core.Response(c, quota, err)
}

// ResetDailyTokens
// @Summary    Reset user's daily token usage
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      uid  path      string  true  "User UID"
// @Success    200  {object}  nil
// @Failure    400  {object}  core.ErrResponse
// @Failure    404  {object}  core.ErrResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/ai/quotas/{uid}/reset-daily [POST].
func (h *QuotaHandler) ResetDailyTokens(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("uid is required"))

		return
	}

	err := h.b.AiQuotas().ResetDailyTokens(c, uid)
	core.Response(c, nil, err)
}
