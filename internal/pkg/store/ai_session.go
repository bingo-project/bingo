// ABOUTME: AI session data access layer.
// ABOUTME: Provides CRUD operations for AI chat sessions.

package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/model"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type AiSessionStore interface {
	Create(ctx context.Context, obj *model.AiSessionM) error
	Update(ctx context.Context, obj *model.AiSessionM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AiSessionM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AiSessionM, error)

	AiSessionExpansion
}

type AiSessionExpansion interface {
	GetBySessionID(ctx context.Context, sessionID string) (*model.AiSessionM, error)
	ListByUID(ctx context.Context, uid string, status string) ([]*model.AiSessionM, error)
	IncrementMessageCount(ctx context.Context, sessionID string, tokens int) error
}

type aiSessionStore struct {
	*genericstore.Store[model.AiSessionM]
}

var _ AiSessionStore = (*aiSessionStore)(nil)

func NewAiSessionStore(store *datastore) *aiSessionStore {
	return &aiSessionStore{
		Store: genericstore.NewStore[model.AiSessionM](store, NewLogger()),
	}
}

func (s *aiSessionStore) GetBySessionID(ctx context.Context, sessionID string) (*model.AiSessionM, error) {
	var session model.AiSessionM
	err := s.DB(ctx).Where("session_id = ?", sessionID).First(&session).Error

	return &session, err
}

func (s *aiSessionStore) ListByUID(ctx context.Context, uid string, status string) ([]*model.AiSessionM, error) {
	var sessions []*model.AiSessionM
	db := s.DB(ctx).Where("uid = ?", uid)
	if status != "" {
		db = db.Where("status = ?", status)
	}
	err := db.Order("updated_at DESC").Find(&sessions).Error

	return sessions, err
}

func (s *aiSessionStore) IncrementMessageCount(ctx context.Context, sessionID string, tokens int) error {
	return s.DB(ctx).
		Model(&model.AiSessionM{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"message_count": gorm.Expr("message_count + 1"),
			"total_tokens":  gorm.Expr("total_tokens + ?", tokens),
		}).Error
}
