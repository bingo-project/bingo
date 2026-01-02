// ABOUTME: AI agent data access layer.
// ABOUTME: Provides CRUD operations for AI agent preset records.

package store

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type AiAgentStore interface {
	Create(ctx context.Context, obj *model.AiAgentM) error
	Update(ctx context.Context, obj *model.AiAgentM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AiAgentM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AiAgentM, error)

	AiAgentExpansion
}

type AiAgentExpansion interface {
	GetByAgentID(ctx context.Context, agentID string) (*model.AiAgentM, error)
	ListByCategory(ctx context.Context, category model.AiAgentCategory, status model.AiAgentStatus) ([]*model.AiAgentM, error)
	ListActive(ctx context.Context) ([]*model.AiAgentM, error)
	FirstOrCreate(ctx context.Context, where *model.AiAgentM, obj *model.AiAgentM) error
}

type aiAgentStore struct {
	*genericstore.Store[model.AiAgentM]
}

var _ AiAgentStore = (*aiAgentStore)(nil)

func NewAiAgentStore(store *datastore) *aiAgentStore {
	return &aiAgentStore{
		Store: genericstore.NewStore[model.AiAgentM](store, NewLogger()),
	}
}

func (s *aiAgentStore) GetByAgentID(ctx context.Context, agentID string) (*model.AiAgentM, error) {
	var agent model.AiAgentM
	err := s.DB(ctx).Where("agent_id = ?", agentID).First(&agent).Error

	return &agent, err
}

func (s *aiAgentStore) ListByCategory(ctx context.Context, category model.AiAgentCategory, status model.AiAgentStatus) ([]*model.AiAgentM, error) {
	var agents []*model.AiAgentM
	db := s.DB(ctx)
	if category != "" {
		db = db.Where("category = ?", category)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	err := db.Order("sort ASC, id ASC").Find(&agents).Error

	return agents, err
}

func (s *aiAgentStore) ListActive(ctx context.Context) ([]*model.AiAgentM, error) {
	var agents []*model.AiAgentM
	err := s.DB(ctx).
		Where("status = ?", model.AiAgentStatusActive).
		Order("sort ASC, id ASC").
		Find(&agents).Error

	return agents, err
}

func (s *aiAgentStore) FirstOrCreate(ctx context.Context, where *model.AiAgentM, obj *model.AiAgentM) error {
	return s.DB(ctx).Where(where).FirstOrCreate(obj).Error
}
