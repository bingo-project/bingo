// ABOUTME: AI role data access layer.
// ABOUTME: Provides CRUD operations for AI role preset records.

package store

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type AiRoleStore interface {
	Create(ctx context.Context, obj *model.AiRoleM) error
	Update(ctx context.Context, obj *model.AiRoleM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AiRoleM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AiRoleM, error)

	AiRoleExpansion
}

type AiRoleExpansion interface {
	GetByRoleID(ctx context.Context, roleID string) (*model.AiRoleM, error)
	ListByCategory(ctx context.Context, category model.AiRoleCategory, status model.AiRoleStatus) ([]*model.AiRoleM, error)
	ListActive(ctx context.Context) ([]*model.AiRoleM, error)
	FirstOrCreate(ctx context.Context, where *model.AiRoleM, obj *model.AiRoleM) error
}

type aiRoleStore struct {
	*genericstore.Store[model.AiRoleM]
}

var _ AiRoleStore = (*aiRoleStore)(nil)

func NewAiRoleStore(store *datastore) *aiRoleStore {
	return &aiRoleStore{
		Store: genericstore.NewStore[model.AiRoleM](store, NewLogger()),
	}
}

func (s *aiRoleStore) GetByRoleID(ctx context.Context, roleID string) (*model.AiRoleM, error) {
	var role model.AiRoleM
	err := s.DB(ctx).Where("role_id = ?", roleID).First(&role).Error

	return &role, err
}

func (s *aiRoleStore) ListByCategory(ctx context.Context, category model.AiRoleCategory, status model.AiRoleStatus) ([]*model.AiRoleM, error) {
	var roles []*model.AiRoleM
	db := s.DB(ctx)
	if category != "" {
		db = db.Where("category = ?", category)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	err := db.Order("sort ASC, id ASC").Find(&roles).Error

	return roles, err
}

func (s *aiRoleStore) ListActive(ctx context.Context) ([]*model.AiRoleM, error) {
	var roles []*model.AiRoleM
	err := s.DB(ctx).
		Where("status = ?", model.AiRoleStatusActive).
		Order("sort ASC, id ASC").
		Find(&roles).Error

	return roles, err
}

func (s *aiRoleStore) FirstOrCreate(ctx context.Context, where *model.AiRoleM, obj *model.AiRoleM) error {
	return s.DB(ctx).Where(where).FirstOrCreate(obj).Error
}
