// ABOUTME: AI provider data access layer.
// ABOUTME: Provides CRUD operations for AI provider configuration.

package store

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type AiProviderStore interface {
	Create(ctx context.Context, obj *model.AiProviderM) error
	Update(ctx context.Context, obj *model.AiProviderM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AiProviderM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AiProviderM, error)

	AiProviderExpansion
}

type AiProviderExpansion interface {
	GetByName(ctx context.Context, name string) (*model.AiProviderM, error)
	ListActive(ctx context.Context) ([]*model.AiProviderM, error)
	GetDefault(ctx context.Context) (*model.AiProviderM, error)
}

type aiProviderStore struct {
	*genericstore.Store[model.AiProviderM]
}

var _ AiProviderStore = (*aiProviderStore)(nil)

func NewAiProviderStore(store *datastore) *aiProviderStore {
	return &aiProviderStore{
		Store: genericstore.NewStore[model.AiProviderM](store, NewLogger()),
	}
}

func (s *aiProviderStore) GetByName(ctx context.Context, name string) (*model.AiProviderM, error) {
	var provider model.AiProviderM
	err := s.DB(ctx).Where("name = ?", name).First(&provider).Error

	return &provider, err
}

func (s *aiProviderStore) ListActive(ctx context.Context) ([]*model.AiProviderM, error) {
	var providers []*model.AiProviderM
	err := s.DB(ctx).
		Where("status = ?", model.AiProviderStatusActive).
		Order("sort ASC, id ASC").
		Find(&providers).Error

	return providers, err
}

func (s *aiProviderStore) GetDefault(ctx context.Context) (*model.AiProviderM, error) {
	var provider model.AiProviderM
	err := s.DB(ctx).
		Where("status = ?", model.AiProviderStatusActive).
		Where("is_default = ?", true).
		First(&provider).Error

	return &provider, err
}
