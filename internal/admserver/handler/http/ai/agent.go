// ABOUTME: HTTP handlers for AI agent management in admin panel.
// ABOUTME: Provides CRUD endpoints for AI agent resources.
package ai

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/admserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type AgentHandler struct {
	b biz.IBiz
}

func NewAgentHandler(ds store.IStore) *AgentHandler {
	return &AgentHandler{b: biz.NewBiz(ds)}
}

// Create
// @Summary    Create AI agent
// @Security   Bearer
// @Tags       AI Agent
// @Accept     application/json
// @Produce    json
// @Param      request  body      v1.CreateAiAgentRequest  true  "Param"
// @Success    200      {object}  v1.AiAgentInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/agents [POST].
func (h *AgentHandler) Create(c *gin.Context) {
	var req v1.CreateAiAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	agent, err := h.b.AiAgents().Create(c, &req)
	core.Response(c, agent, err)
}

// Get
// @Summary    Get AI agent by agent_id
// @Security   Bearer
// @Tags       AI Agent
// @Accept     application/json
// @Produce    json
// @Param      id  path      string  true  "Agent ID"
// @Success    200  {object}  v1.AiAgentInfo
// @Failure    400  {object}  core.ErrResponse
// @Failure    404  {object}  core.ErrResponse
// @Router     /v1/ai/agents/{id} [GET].
func (h *AgentHandler) Get(c *gin.Context) {
	agentID := c.Param("id")
	agent, err := h.b.AiAgents().Get(c, agentID)
	core.Response(c, agent, err)
}

// List
// @Summary    List AI agents
// @Security   Bearer
// @Tags       AI Agent
// @Accept     application/json
// @Produce    json
// @Param      category  query     string  false  "Filter by category"
// @Param      status    query     string  false  "Filter by status"
// @Success    200       {object}  v1.ListAiAgentResponse
// @Failure    400       {object}  core.ErrResponse
// @Router     /v1/ai/agents [GET].
func (h *AgentHandler) List(c *gin.Context) {
	var req v1.ListAiAgentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	agents, err := h.b.AiAgents().List(c, &req)
	core.Response(c, agents, err)
}

// Update
// @Summary    Update AI agent
// @Security   Bearer
// @Tags       AI Agent
// @Accept     application/json
// @Produce    json
// @Param      id        path      string                   true  "Agent ID"
// @Param      request  body      v1.UpdateAiAgentRequest   true  "Update request"
// @Success    200      {object}  v1.AiAgentInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    404      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/agents/{id} [PUT].
func (h *AgentHandler) Update(c *gin.Context) {
	var req v1.UpdateAiAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	agentID := c.Param("id")
	agent, err := h.b.AiAgents().Update(c, agentID, &req)
	core.Response(c, agent, err)
}

// Delete
// @Summary    Delete AI agent
// @Security   Bearer
// @Tags       AI Agent
// @Accept     application/json
// @Produce    json
// @Param      id  path  string  true  "Agent ID"
// @Success    200  {object}  nil
// @Failure    404  {object}  core.ErrResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/ai/agents/{id} [DELETE].
func (h *AgentHandler) Delete(c *gin.Context) {
	agentID := c.Param("id")
	err := h.b.AiAgents().Delete(c, agentID)
	core.Response(c, nil, err)
}
