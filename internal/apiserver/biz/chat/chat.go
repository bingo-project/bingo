// ABOUTME: Chat business logic interface and implementation.
// ABOUTME: Orchestrates AI chat, session management, and provider selection.

package chat

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ai"
)

// ChatBiz defines the chat business interface
type ChatBiz interface {
	// Chat performs a non-streaming chat completion
	Chat(ctx context.Context, uid string, req *ai.ChatRequest) (*ai.ChatResponse, error)

	// ChatStream performs a streaming chat completion
	ChatStream(ctx context.Context, uid string, req *ai.ChatRequest) (*ai.ChatStream, error)

	// Sessions returns the session management interface
	Sessions() SessionBiz

	// ListModels returns available models
	ListModels(ctx context.Context) ([]ai.ModelInfo, error)
}

type chatBiz struct {
	ds       store.IStore
	registry *ai.Registry
}

var _ ChatBiz = (*chatBiz)(nil)

// New creates a new ChatBiz instance
func New(ds store.IStore, registry *ai.Registry) *chatBiz {
	return &chatBiz{
		ds:       ds,
		registry: registry,
	}
}

func (b *chatBiz) Chat(ctx context.Context, uid string, req *ai.ChatRequest) (*ai.ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, errno.ErrAIEmptyMessages
	}

	// Get provider for the model
	provider, ok := b.registry.GetByModel(req.Model)
	if !ok {
		return nil, errno.ErrAIModelNotFound
	}

	// Call provider
	resp, err := provider.Chat(ctx, req)
	if err != nil {
		return nil, errno.ErrAIProviderError.WithMessage("chat failed: %v", err)
	}

	// Save to session if session ID provided
	if req.SessionID != "" {
		go b.saveToSession(context.Background(), req, resp)
	}

	return resp, nil
}

func (b *chatBiz) ChatStream(ctx context.Context, uid string, req *ai.ChatRequest) (*ai.ChatStream, error) {
	if len(req.Messages) == 0 {
		return nil, errno.ErrAIEmptyMessages
	}

	// Get provider for the model
	provider, ok := b.registry.GetByModel(req.Model)
	if !ok {
		return nil, errno.ErrAIModelNotFound
	}

	// Call provider
	stream, err := provider.ChatStream(ctx, req)
	if err != nil {
		return nil, errno.ErrAIProviderError.WithMessage("stream failed: %v", err)
	}

	return stream, nil
}

func (b *chatBiz) Sessions() SessionBiz {
	return NewSession(b.ds)
}

func (b *chatBiz) ListModels(ctx context.Context) ([]ai.ModelInfo, error) {
	return b.registry.ListModels(), nil
}

// saveToSession saves request and response to session (background goroutine)
func (b *chatBiz) saveToSession(ctx context.Context, req *ai.ChatRequest, resp *ai.ChatResponse) {
	// Save user message
	for _, msg := range req.Messages {
		if msg.Role == ai.RoleUser {
			_ = b.ds.AiMessage().Create(ctx, &model.AiMessageM{
				SessionID: req.SessionID,
				Role:      msg.Role,
				Content:   msg.Content,
				Model:     req.Model,
			})
		}
	}

	// Save assistant response
	if len(resp.Choices) > 0 {
		_ = b.ds.AiMessage().Create(ctx, &model.AiMessageM{
			SessionID: req.SessionID,
			Role:      ai.RoleAssistant,
			Content:   resp.Choices[0].Message.Content,
			Tokens:    resp.Usage.CompletionTokens,
			Model:     resp.Model,
		})
	}

	// Update session stats
	_ = b.ds.AiSession().IncrementMessageCount(ctx, req.SessionID, resp.Usage.TotalTokens)
}
