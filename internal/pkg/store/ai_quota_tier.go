// ABOUTME: AI quota tier data access layer.
// ABOUTME: Provides CRUD operations for quota tier definitions.

package store

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type AiQuotaTierStore interface {
	Create(ctx context.Context, obj *model.AiQuotaTierM) error
	Update(ctx context.Context, obj *model.AiQuotaTierM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AiQuotaTierM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AiQuotaTierM, error)

	AiQuotaTierExpansion
}

type AiQuotaTierExpansion interface {
	GetByTier(ctx context.Context, tier string) (*model.AiQuotaTierM, error)
}

type aiQuotaTierStore struct {
	*genericstore.Store[model.AiQuotaTierM]
}

var _ AiQuotaTierStore = (*aiQuotaTierStore)(nil)

func NewAiQuotaTierStore(store *datastore) *aiQuotaTierStore {
	return &aiQuotaTierStore{
		Store: genericstore.NewStore[model.AiQuotaTierM](store, NewLogger()),
	}
}

func (s *aiQuotaTierStore) GetByTier(ctx context.Context, tier string) (*model.AiQuotaTierM, error) {
	var t model.AiQuotaTierM
	err := s.DB(ctx).Where("tier = ?", tier).First(&t).Error

	return &t, err
}
