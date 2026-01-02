// ABOUTME: AI role business logic for admin management.
// ABOUTME: Provides full CRUD operations for AI role presets with no status restrictions.
package ai

import (
	"context"
	"errors"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

// AiAgentBiz defines AI agent management interface for admin.
type AiAgentBiz interface {
	Create(ctx context.Context, req *v1.CreateAiAgentRequest) (*v1.AiAgentInfo, error)
	Get(ctx context.Context, agentID string) (*v1.AiAgentInfo, error)
	List(ctx context.Context, req *v1.ListAiAgentRequest) (*v1.ListAiAgentResponse, error)
	Update(ctx context.Context, agentID string, req *v1.UpdateAiAgentRequest) (*v1.AiAgentInfo, error)
	Delete(ctx context.Context, agentID string) error
}

type aiRoleBiz struct {
	ds store.IStore
}

type aiAgentBiz struct {
	ds store.IStore
}

var _ AiAgentBiz = (*aiAgentBiz)(nil)

func NewAiAgent(ds store.IStore) AiAgentBiz {
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

func (b *aiAgentBiz) Create(ctx context.Context, req *v1.CreateAiAgentRequest) (*v1.AiAgentInfo, error) {
	// Check if agent already exists
	existing, err := b.ds.AiAgents().GetByAgentID(ctx, req.AgentID)
	if err == nil && existing != nil {
		return nil, errno.ErrResourceAlreadyExists.WithMessage("agent_id already exists: %s", req.AgentID)
	}

	// Set default category if not provided
	category := model.AiAgentCategoryGeneral
	if req.Category != "" {
		category = model.AiAgentCategory(req.Category)
	}

	agent := &model.AiAgentM{
		AgentID:      req.AgentID,
		Name:         req.Name,
		Description:  req.Description,
		Icon:         req.Icon,
		Category:     category,
		SystemPrompt: req.SystemPrompt,
		Model:        req.Model,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		Sort:         req.Sort,
		Status:       model.AiAgentStatusActive,
	}

	if err := b.ds.AiAgents().Create(ctx, agent); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("create ai agent: %v", err)
	}

	log.C(ctx).Infow("ai agent created", "agent_id", agent.AgentID, "name", agent.Name)

	return toAgentInfo(agent), nil
}

func (b *aiAgentBiz) Get(ctx context.Context, agentID string) (*v1.AiAgentInfo, error) {
	agent, err := b.ds.AiAgents().GetByAgentID(ctx, agentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIRoleNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("get ai agent: %v", err)
	}

	return toAgentInfo(agent), nil
}

func (b *aiAgentBiz) List(ctx context.Context, req *v1.ListAiAgentRequest) (*v1.ListAiAgentResponse, error) {
	var category model.AiAgentCategory
	if req.Category != "" {
		category = model.AiAgentCategory(req.Category)
	}

	// Admin can list all statuses, no default filtering
	var status model.AiAgentStatus
	if req.Status != "" {
		status = model.AiAgentStatus(req.Status)
	}

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

func (b *aiAgentBiz) Update(ctx context.Context, agentID string, req *v1.UpdateAiAgentRequest) (*v1.AiAgentInfo, error) {
	agent, err := b.ds.AiAgents().GetByAgentID(ctx, agentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIRoleNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("get ai agent: %v", err)
	}

	// Update fields
	_ = copier.CopyWithOption(agent, req, copier.Option{IgnoreEmpty: true})
	if req.Category != "" {
		agent.Category = model.AiAgentCategory(req.Category)
	}
	if req.Status != "" {
		agent.Status = model.AiAgentStatus(req.Status)
	}

	if err := b.ds.AiAgents().Update(ctx, agent); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("update ai agent: %v", err)
	}

	log.C(ctx).Infow("ai agent updated", "agent_id", agent.AgentID)

	return toAgentInfo(agent), nil
}

func (b *aiAgentBiz) Delete(ctx context.Context, agentID string) error {
	agent, err := b.ds.AiAgents().GetByAgentID(ctx, agentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrAIRoleNotFound
		}

		return errno.ErrDBRead.WithMessage("get ai agent: %v", err)
	}

	agent.Status = model.AiAgentStatusDisabled
	if err := b.ds.AiAgents().Update(ctx, agent, "status"); err != nil {
		return errno.ErrDBWrite.WithMessage("delete ai agent: %v", err)
	}

	log.C(ctx).Infow("ai agent deleted", "agent_id", agent.AgentID)

	return nil
}
