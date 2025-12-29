// ABOUTME: AI message data access layer.
// ABOUTME: Provides CRUD operations for chat messages.

package store

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type AiMessageStore interface {
	Create(ctx context.Context, obj *model.AiMessageM) error
	Update(ctx context.Context, obj *model.AiMessageM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AiMessageM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AiMessageM, error)

	AiMessageExpansion
}

type AiMessageExpansion interface {
	ListBySessionID(ctx context.Context, sessionID string, limit int) ([]*model.AiMessageM, error)
	DeleteBySessionID(ctx context.Context, sessionID string) error
}

type aiMessageStore struct {
	*genericstore.Store[model.AiMessageM]
}

var _ AiMessageStore = (*aiMessageStore)(nil)

func NewAiMessageStore(store *datastore) *aiMessageStore {
	return &aiMessageStore{
		Store: genericstore.NewStore[model.AiMessageM](store, NewLogger()),
	}
}

func (s *aiMessageStore) ListBySessionID(ctx context.Context, sessionID string, limit int) ([]*model.AiMessageM, error) {
	var messages []*model.AiMessageM
	db := s.DB(ctx).Where("session_id = ?", sessionID).Order("created_at ASC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	err := db.Find(&messages).Error

	return messages, err
}

func (s *aiMessageStore) DeleteBySessionID(ctx context.Context, sessionID string) error {
	return s.DB(ctx).Where("session_id = ?", sessionID).Delete(&model.AiMessageM{}).Error
}
