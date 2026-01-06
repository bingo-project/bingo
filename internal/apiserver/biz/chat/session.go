// ABOUTME: Session business logic implementation.
// ABOUTME: Manages AI chat session lifecycle and history.

package chat

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ai"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

// SessionBiz defines session management interface
type SessionBiz interface {
	Create(ctx context.Context, uid string, title string, modelName string, agentID string) (*v1.SessionInfo, error)
	Get(ctx context.Context, uid string, sessionID string) (*v1.SessionInfo, error)
	List(ctx context.Context, uid string) ([]v1.SessionInfo, error)
	Update(ctx context.Context, uid string, sessionID string, title string, modelName string) (*v1.SessionInfo, error)
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

// toSessionInfo converts model.AiSessionM to v1.SessionInfo
func toSessionInfo(m *model.AiSessionM) *v1.SessionInfo {
	var info v1.SessionInfo
	_ = copier.Copy(&info, m)
	info.SessionID = m.SessionID

	return &info
}

func (b *sessionBiz) Create(ctx context.Context, uid string, title string, modelName string, agentID string) (*v1.SessionInfo, error) {
	var selectedModel string
	var finalTitle string

	// 如果指定了 agent_id, 从角色获取默认配置
	if agentID != "" {
		agent, err := b.ds.AiAgents().GetByAgentID(ctx, agentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errno.ErrAIRoleNotFound
			}

			return nil, errno.ErrDBRead.WithMessage("get ai agent: %v", err)
		}
		if agent.Status == model.AiAgentStatusDisabled {
			return nil, errno.ErrAIRoleDisabled
		}
		selectedModel = agent.Model // 使用角色的模型
		if title == "" {
			finalTitle = agent.Name // 默认标题使用角色名称
		}
	}

	// 使用请求的模型覆盖角色模型(如果提供)
	if modelName != "" {
		selectedModel = modelName
	}

	// 回退到默认模型
	if selectedModel == "" {
		selectedModel = "gpt-4o" // TODO: 从配置读取
	}
	if finalTitle == "" {
		if title != "" {
			finalTitle = title
		} else {
			finalTitle = "新对话"
		}
	}

	session := &model.AiSessionM{
		SessionID: uuid.NewString(),
		UID:       uid,
		AgentID:   agentID, // 绑定角色
		Title:     finalTitle,
		Model:     selectedModel,
		Status:    model.AiSessionStatusActive,
	}

	if err := b.ds.AiSession().Create(ctx, session); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("create session: %v", err)
	}

	return toSessionInfo(session), nil
}

func (b *sessionBiz) Get(ctx context.Context, uid string, sessionID string) (*v1.SessionInfo, error) {
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

	return toSessionInfo(session), nil
}

func (b *sessionBiz) List(ctx context.Context, uid string) ([]v1.SessionInfo, error) {
	sessions, err := b.ds.AiSession().ListByUID(ctx, uid, model.AiSessionStatusActive)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("list sessions: %v", err)
	}

	result := make([]v1.SessionInfo, len(sessions))
	for i, s := range sessions {
		result[i] = *toSessionInfo(s)
	}

	return result, nil
}

func (b *sessionBiz) Update(ctx context.Context, uid string, sessionID string, title string, modelName string) (*v1.SessionInfo, error) {
	// Fetch model for update
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

	return toSessionInfo(session), nil
}

func (b *sessionBiz) Delete(ctx context.Context, uid string, sessionID string) error {
	// Fetch model for update
	session, err := b.ds.AiSession().GetBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrAISessionNotFound
		}

		return errno.ErrDBRead.WithMessage("get session: %v", err)
	}

	// Check ownership
	if session.UID != uid {
		return errno.ErrAISessionNotFound
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
