// ABOUTME: Session business logic implementation.
// ABOUTME: Manages AI chat session lifecycle and history.

package chat

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ai"
)

// SessionBiz defines session management interface
type SessionBiz interface {
	Create(ctx context.Context, uid string, title string, model string) (*model.AiSessionM, error)
	Get(ctx context.Context, uid string, sessionID string) (*model.AiSessionM, error)
	List(ctx context.Context, uid string) ([]*model.AiSessionM, error)
	Update(ctx context.Context, uid string, sessionID string, title string, modelName string) (*model.AiSessionM, error)
	Delete(ctx context.Context, uid string, sessionID string) error
	GetHistory(ctx context.Context, sessionID string, limit int) ([]ai.Message, error)
}

type sessionBiz struct {
	ds store.IStore
}

var _ SessionBiz = (*sessionBiz)(nil)

func NewSession(ds store.IStore) *sessionBiz {
	return &sessionBiz{ds: ds}
}

func (b *sessionBiz) Create(ctx context.Context, uid string, title string, modelName string) (*model.AiSessionM, error) {
	session := &model.AiSessionM{
		SessionID: uuid.NewString(),
		UID:       uid,
		Title:     title,
		Model:     modelName,
		Status:    model.AiSessionStatusActive,
	}

	if err := b.ds.AiSession().Create(ctx, session); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("create session: %v", err)
	}

	return session, nil
}

func (b *sessionBiz) Get(ctx context.Context, uid string, sessionID string) (*model.AiSessionM, error) {
	session, err := b.ds.AiSession().GetBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAISessionNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("get session: %v", err)
	}

	// Check ownership
	if session.UID != uid {
		return nil, errno.ErrAISessionNotFound
	}

	return session, nil
}

func (b *sessionBiz) List(ctx context.Context, uid string) ([]*model.AiSessionM, error) {
	sessions, err := b.ds.AiSession().ListByUID(ctx, uid, model.AiSessionStatusActive)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("list sessions: %v", err)
	}

	return sessions, nil
}

func (b *sessionBiz) Update(ctx context.Context, uid string, sessionID string, title string, modelName string) (*model.AiSessionM, error) {
	session, err := b.Get(ctx, uid, sessionID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	var fields []string
	if title != "" {
		session.Title = title
		fields = append(fields, "title")
	}
	if modelName != "" {
		session.Model = modelName
		fields = append(fields, "model")
	}

	if len(fields) > 0 {
		if err := b.ds.AiSession().Update(ctx, session, fields...); err != nil {
			return nil, errno.ErrDBWrite.WithMessage("update session: %v", err)
		}
	}

	return session, nil
}

func (b *sessionBiz) Delete(ctx context.Context, uid string, sessionID string) error {
	session, err := b.Get(ctx, uid, sessionID)
	if err != nil {
		return err
	}

	session.Status = model.AiSessionStatusDeleted
	if err := b.ds.AiSession().Update(ctx, session, "status"); err != nil {
		return errno.ErrDBWrite.WithMessage("delete session: %v", err)
	}

	return nil
}

func (b *sessionBiz) GetHistory(ctx context.Context, sessionID string, limit int) ([]ai.Message, error) {
	messages, err := b.ds.AiMessage().ListBySessionID(ctx, sessionID, limit)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("get history: %v", err)
	}

	result := make([]ai.Message, len(messages))
	for i, m := range messages {
		result[i] = ai.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	return result, nil
}
