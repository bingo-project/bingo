// ABOUTME: Chat business logic interface and implementation.
// ABOUTME: Orchestrates AI chat, session management, and provider selection.

package chat

import (
	"context"
	"time"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ai"
)

const (
	// saveSessionTimeout is the timeout for background session save operations
	saveSessionTimeout = 30 * time.Second
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
	quota    *quotaChecker
}

var _ ChatBiz = (*chatBiz)(nil)

// New creates a new ChatBiz instance
func New(ds store.IStore, registry *ai.Registry) *chatBiz {
	return &chatBiz{
		ds:       ds,
		registry: registry,
		quota:    newQuotaChecker(ds),
	}
}

func (b *chatBiz) Chat(ctx context.Context, uid string, req *ai.ChatRequest) (*ai.ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, errno.ErrAIEmptyMessages
	}

	// Check TPD quota before calling provider
	if err := b.quota.CheckTPD(ctx, uid); err != nil {
		return nil, err
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

	// Update TPD quota after successful response (background)
	if resp.Usage.TotalTokens > 0 {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
			defer cancel()
			if err := b.quota.UpdateTPD(ctx, uid, resp.Usage.TotalTokens); err != nil {
				log.C(ctx).Errorw("Failed to update TPD quota", "uid", uid, "tokens", resp.Usage.TotalTokens, "err", err)
			}
		}()
	}

	// Save to session if session ID provided (background with timeout)
	if req.SessionID != "" {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
			defer cancel()
			b.saveToSession(ctx, uid, req, resp)
		}()
	}

	return resp, nil
}

func (b *chatBiz) ChatStream(ctx context.Context, uid string, req *ai.ChatRequest) (*ai.ChatStream, error) {
	if len(req.Messages) == 0 {
		return nil, errno.ErrAIEmptyMessages
	}

	// Check TPD quota before calling provider
	if err := b.quota.CheckTPD(ctx, uid); err != nil {
		return nil, err
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

	// Wrap stream to save messages after completion
	return b.wrapStreamForSaving(stream, uid, req), nil
}

// wrapStreamForSaving wraps a stream to save messages after completion.
func (b *chatBiz) wrapStreamForSaving(stream *ai.ChatStream, uid string, req *ai.ChatRequest) *ai.ChatStream {
	wrapped := ai.NewChatStream(100)

	go func() {
		var contentBuilder []byte
		var modelName string
		var totalTokens int

		for {
			chunk, err := stream.Recv()
			if err != nil {
				// Stream ended, save accumulated content
				if len(contentBuilder) > 0 && req.SessionID != "" {
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
						defer cancel()
						b.saveStreamToSession(ctx, uid, req, string(contentBuilder), modelName, totalTokens)
					}()
				}
				// Update TPD quota (estimate if not available)
				if totalTokens > 0 {
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
						defer cancel()
						if err := b.quota.UpdateTPD(ctx, uid, totalTokens); err != nil {
							log.C(ctx).Errorw("Failed to update TPD quota", "uid", uid, "tokens", totalTokens, "err", err)
						}
					}()
				}
				wrapped.CloseWithError(err)

				return
			}

			// Accumulate content
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
				contentBuilder = append(contentBuilder, chunk.Choices[0].Delta.Content...)
			}
			if chunk.Model != "" {
				modelName = chunk.Model
			}
			if chunk.Usage != nil {
				totalTokens = chunk.Usage.TotalTokens
			}

			wrapped.Send(chunk)
		}
	}()

	return wrapped
}

// saveStreamToSession saves stream messages to session.
func (b *chatBiz) saveStreamToSession(ctx context.Context, uid string, req *ai.ChatRequest, content string, modelName string, tokens int) {
	// Save user message
	for _, msg := range req.Messages {
		if msg.Role == ai.RoleUser {
			if err := b.ds.AiMessage().Create(ctx, &model.AiMessageM{
				SessionID: req.SessionID,
				Role:      msg.Role,
				Content:   msg.Content,
				Model:     req.Model,
			}); err != nil {
				log.C(ctx).Errorw("Failed to save user message", "session_id", req.SessionID, "uid", uid, "err", err)
			}
		}
	}

	// Save assistant response
	if content != "" {
		usedModel := modelName
		if usedModel == "" {
			usedModel = req.Model
		}
		if err := b.ds.AiMessage().Create(ctx, &model.AiMessageM{
			SessionID: req.SessionID,
			Role:      ai.RoleAssistant,
			Content:   content,
			Tokens:    tokens,
			Model:     usedModel,
		}); err != nil {
			log.C(ctx).Errorw("Failed to save assistant message", "session_id", req.SessionID, "uid", uid, "err", err)
		}
	}

	// Update session stats
	if err := b.ds.AiSession().IncrementMessageCount(ctx, req.SessionID, tokens); err != nil {
		log.C(ctx).Errorw("Failed to update session stats", "session_id", req.SessionID, "uid", uid, "err", err)
	}
}

func (b *chatBiz) Sessions() SessionBiz {
	return NewSession(b.ds)
}

func (b *chatBiz) ListModels(ctx context.Context) ([]ai.ModelInfo, error) {
	return b.registry.ListModels(), nil
}

// saveToSession saves request and response to session (background goroutine)
func (b *chatBiz) saveToSession(ctx context.Context, uid string, req *ai.ChatRequest, resp *ai.ChatResponse) {
	// Save user message
	for _, msg := range req.Messages {
		if msg.Role == ai.RoleUser {
			if err := b.ds.AiMessage().Create(ctx, &model.AiMessageM{
				SessionID: req.SessionID,
				Role:      msg.Role,
				Content:   msg.Content,
				Model:     req.Model,
			}); err != nil {
				log.C(ctx).Errorw("Failed to save user message", "session_id", req.SessionID, "uid", uid, "err", err)
			}
		}
	}

	// Save assistant response
	if len(resp.Choices) > 0 {
		if err := b.ds.AiMessage().Create(ctx, &model.AiMessageM{
			SessionID: req.SessionID,
			Role:      ai.RoleAssistant,
			Content:   resp.Choices[0].Message.Content,
			Tokens:    resp.Usage.CompletionTokens,
			Model:     resp.Model,
		}); err != nil {
			log.C(ctx).Errorw("Failed to save assistant message", "session_id", req.SessionID, "uid", uid, "err", err)
		}
	}

	// Update session stats
	if err := b.ds.AiSession().IncrementMessageCount(ctx, req.SessionID, resp.Usage.TotalTokens); err != nil {
		log.C(ctx).Errorw("Failed to update session stats", "session_id", req.SessionID, "uid", uid, "err", err)
	}
}
