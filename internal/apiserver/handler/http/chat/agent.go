// ABOUTME: HTTP handlers for AI agent.
// ABOUTME: Provides read-only access to available agents. preset resources.

package chat

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/apiserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type AgentHandler struct {
	b biz.IBiz
}

func NewAgentHandler(ds store.IStore) *AgentHandler {
	return &AgentHandler{
		b: biz.NewBiz(ds),
	}
}

// Get
// @Summary    Get AI agent details
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      id  path      string  true  "Agent ID"
// @Success    200  {object}  v1.AiAgentInfo
// @Failure    404  {object}  core.ErrResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/ai/agents/{id} [GET].
func (h *AgentHandler) Get(c *gin.Context) {
	agentID := c.Param("id")
	agent, err := h.b.AiAgents().Get(c, agentID)
	core.Response(c, agent, err)
}

// List
// @Summary    List available AI agents
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      category  query     string  false  "Filter by category"
// @Success    200       {object}  v1.ListAiAgentResponse
// @Failure    500       {object}  core.ErrResponse
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
