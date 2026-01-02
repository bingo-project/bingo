// ABOUTME: AI role business logic implementation.
// ABOUTME: Handles role query operations and validations.

package chat

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

// AiAgentBiz defines AI agent query interface for users.
type AiAgentBiz interface {
	Get(ctx context.Context, agentID string) (*v1.AiAgentInfo, error)
	List(ctx context.Context, req *v1.ListAiAgentRequest) (*v1.ListAiAgentResponse, error)
}

type aiAgentBiz struct {
	ds store.IStore
}

var _ AiAgentBiz = (*aiAgentBiz)(nil)

func NewAiAgent(ds store.IStore) *aiAgentBiz {
	return &aiAgentBiz{ds: ds}
}

// toAgentInfo converts model.AiAgentM to v1.AiAgentInfo.
func toAgentInfo(m *model.AiAgentM) *v1.AiAgentInfo {
	return &v1.AiAgentInfo{
		AgentID:      m.AgentID,
		Name:         m.Name,
		Description:  m.Description,
		Icon:         m.Icon,
		Category:     string(m.Category),
		SystemPrompt: m.SystemPrompt,
		Model:        m.Model,
		Temperature:  m.Temperature,
		MaxTokens:    m.MaxTokens,
		Sort:         m.Sort,
		Status:       string(m.Status),
	}
}

func (b *aiAgentBiz) Get(ctx context.Context, agentID string) (*v1.AiAgentInfo, error) {
	agent, err := b.ds.AiAgents().GetByAgentID(ctx, agentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIRoleNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("get ai agent: %v", err)
	}

	// Only return active agents
	if agent.Status != model.AiAgentStatusActive {
		return nil, errno.ErrAIRoleNotFound
	}

	return toAgentInfo(agent), nil
}

func (b *aiAgentBiz) List(ctx context.Context, req *v1.ListAiAgentRequest) (*v1.ListAiAgentResponse, error) {
	var category model.AiAgentCategory
	if req.Category != "" {
		category = model.AiAgentCategory(req.Category)
	}

	// Force filter to active only, ignore status parameter
	status := model.AiAgentStatusActive

	agents, err := b.ds.AiAgents().ListByCategory(ctx, category, status)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("list ai agents: %v", err)
	}

	data := make([]v1.AiAgentInfo, len(agents))
	for i, r := range agents {
		data[i] = *toAgentInfo(r)
	}

	return &v1.ListAiAgentResponse{
		Total: int64(len(agents)),
		Data:  data,
	}, nil
}
