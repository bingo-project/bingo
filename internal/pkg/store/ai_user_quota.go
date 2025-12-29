// ABOUTME: AI user quota data access layer.
// ABOUTME: Provides CRUD operations for per-user quota tracking.

package store

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/model"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type AiUserQuotaStore interface {
	Create(ctx context.Context, obj *model.AiUserQuotaM) error
	Update(ctx context.Context, obj *model.AiUserQuotaM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AiUserQuotaM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AiUserQuotaM, error)

	AiUserQuotaExpansion
}

type AiUserQuotaExpansion interface {
	GetByUID(ctx context.Context, uid string) (*model.AiUserQuotaM, error)
	IncrementTokens(ctx context.Context, uid string, tokens int) error
	ResetDailyTokens(ctx context.Context, uid string) error
}

type aiUserQuotaStore struct {
	*genericstore.Store[model.AiUserQuotaM]
}

var _ AiUserQuotaStore = (*aiUserQuotaStore)(nil)

func NewAiUserQuotaStore(store *datastore) *aiUserQuotaStore {
	return &aiUserQuotaStore{
		Store: genericstore.NewStore[model.AiUserQuotaM](store, NewLogger()),
	}
}

func (s *aiUserQuotaStore) GetByUID(ctx context.Context, uid string) (*model.AiUserQuotaM, error) {
	var quota model.AiUserQuotaM
	err := s.DB(ctx).Where("uid = ?", uid).First(&quota).Error

	return &quota, err
}

func (s *aiUserQuotaStore) IncrementTokens(ctx context.Context, uid string, tokens int) error {
	return s.DB(ctx).
		Model(&model.AiUserQuotaM{}).
		Where("uid = ?", uid).
		Update("used_tokens_today", gorm.Expr("used_tokens_today + ?", tokens)).
		Error
}

func (s *aiUserQuotaStore) ResetDailyTokens(ctx context.Context, uid string) error {
	now := time.Now()

	return s.DB(ctx).
		Model(&model.AiUserQuotaM{}).
		Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"used_tokens_today": 0,
			"last_reset_at":     &now,
		}).Error
}
