// ABOUTME: AI model data access layer.
// ABOUTME: Provides CRUD operations for AI model configuration.

package store

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type AiModelStore interface {
	Create(ctx context.Context, obj *model.AiModelM) error
	Update(ctx context.Context, obj *model.AiModelM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AiModelM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AiModelM, error)

	AiModelExpansion
}

type AiModelExpansion interface {
	GetByProviderAndModel(ctx context.Context, providerName, model string) (*model.AiModelM, error)
	FindActiveByModel(ctx context.Context, model string) (*model.AiModelM, error)
	ListByProvider(ctx context.Context, providerName string) ([]*model.AiModelM, error)
	ListActive(ctx context.Context) ([]*model.AiModelM, error)
	GetDefault(ctx context.Context) (*model.AiModelM, error)
	FirstOrCreate(ctx context.Context, where *model.AiModelM, obj *model.AiModelM) error
}

type aiModelStore struct {
	*genericstore.Store[model.AiModelM]
}

var _ AiModelStore = (*aiModelStore)(nil)

func NewAiModelStore(store *datastore) *aiModelStore {
	return &aiModelStore{
		Store: genericstore.NewStore[model.AiModelM](store, NewLogger()),
	}
}

func (s *aiModelStore) GetByProviderAndModel(ctx context.Context, providerName, modelID string) (*model.AiModelM, error) {
	var m model.AiModelM
	err := s.DB(ctx).Where("provider_name = ? AND model = ?", providerName, modelID).First(&m).Error

	return &m, err
}

func (s *aiModelStore) FindActiveByModel(ctx context.Context, modelID string) (*model.AiModelM, error) {
	var m model.AiModelM
	err := s.DB(ctx).
		Where("model = ?", modelID).
		Where("status = ?", model.AiModelStatusActive).
		Order("sort ASC, id ASC").
		First(&m).Error

	return &m, err
}

func (s *aiModelStore) ListByProvider(ctx context.Context, providerName string) ([]*model.AiModelM, error) {
	var models []*model.AiModelM
	err := s.DB(ctx).
		Where("provider_name = ?", providerName).
		Where("status = ?", model.AiModelStatusActive).
		Order("sort ASC, id ASC").
		Find(&models).Error

	return models, err
}

func (s *aiModelStore) ListActive(ctx context.Context) ([]*model.AiModelM, error) {
	var models []*model.AiModelM
	err := s.DB(ctx).
		Where("status = ?", model.AiModelStatusActive).
		Order("sort ASC, id ASC").
		Find(&models).Error

	return models, err
}

func (s *aiModelStore) GetDefault(ctx context.Context) (*model.AiModelM, error) {
	var m model.AiModelM
	err := s.DB(ctx).
		Where("status = ?", model.AiModelStatusActive).
		Where("is_default = ?", true).
		First(&m).Error

	return &m, err
}

func (s *aiModelStore) FirstOrCreate(ctx context.Context, where *model.AiModelM, obj *model.AiModelM) error {
	return s.DB(ctx).Where(where).FirstOrCreate(obj).Error
}
