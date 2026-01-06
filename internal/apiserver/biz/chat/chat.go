// ABOUTME: Chat business logic interface and implementation.
// ABOUTME: Orchestrates AI chat, session management, and provider selection.

package chat

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/ai"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	aipkg "github.com/bingo-project/bingo/pkg/ai"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

const (
	// saveSessionTimeout is the timeout for background session save operations
	saveSessionTimeout = 30 * time.Second
)

// ChatBiz defines the chat business interface
type ChatBiz interface {
	// Chat performs a non-streaming chat completion
	Chat(ctx context.Context, uid string, req *aipkg.ChatRequest) (*aipkg.ChatResponse, error)

	// ChatStream performs a streaming chat completion
	ChatStream(ctx context.Context, uid string, req *aipkg.ChatRequest) (*aipkg.ChatStream, error)

	// Sessions returns the session management interface
	Sessions() SessionBiz

	// ListModels returns available models (OpenAI-compatible format)
	ListModels(ctx context.Context) (*v1.ListModelsResponse, error)
}

type chatBiz struct {
	ds       store.IStore
	registry *aipkg.Registry
	quota    *quotaChecker
	fallback *ai.FallbackSelector
}

var _ ChatBiz = (*chatBiz)(nil)

// New creates a new ChatBiz instance
func New(ds store.IStore, registry *aipkg.Registry) *chatBiz {
	return &chatBiz{
		ds:       ds,
		registry: registry,
		quota:    newQuotaChecker(ds),
		fallback: ai.NewFallbackSelector(ds.AiModel(), registry),
	}
}

func (b *chatBiz) Chat(ctx context.Context, uid string, req *aipkg.ChatRequest) (*aipkg.ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, errno.ErrAIEmptyMessages
	}

	// Validate session if provided and get role_id from session
	if req.SessionID != "" {
		if err := b.validateSession(ctx, req.SessionID, uid); err != nil {
			return nil, err
		}

		// Get agent_id from session
		session, err := b.ds.AiSession().GetBySessionID(ctx, req.SessionID)
		if err == nil && session.AgentID != "" {
			req.AgentID = session.AgentID // Use session's agent
		}
	}

	// Apply agent preset if specified
	if err := b.buildMessagesWithAgent(ctx, req); err != nil {
		return nil, err
	}

	// Resolve model (request > session > default)
	req.Model = b.resolveModel(ctx, req.Model, req.SessionID)
	if req.Model == "" {
		return nil, errno.ErrAIModelNotFound
	}

	// Capture new messages BEFORE loading history
	newMessages := req.Messages

	// Load and merge history messages
	messages, err := b.loadAndMergeHistory(ctx, req.SessionID, req.Messages)
	if err != nil {
		return nil, err
	}
	req.Messages = messages

	// Reserve TPD quota atomically before calling provider
	reservedTokens, err := b.quota.ReserveTPD(ctx, uid, req.MaxTokens)
	if err != nil {
		return nil, err
	}

	// Ensure quota is released if not consumed (defer pattern)
	quotaConsumed := false
	defer func() {
		if !quotaConsumed && reservedTokens > 0 {
			// Release in background to avoid blocking
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
				defer cancel()
				if err := b.quota.AdjustTPD(ctx, uid, 0, reservedTokens); err != nil {
					log.C(ctx).Errorw("Failed to release reserved quota",
						"uid", uid, "reserved", reservedTokens, "err", err)
				}
			}()
		}
	}()

	// Get provider with fallback
	provider, modelUsed, err := b.getProviderWithFallback(ctx, req.Model)
	if err != nil {
		return nil, err
	}
	req.Model = modelUsed

	// Call provider
	resp, err := provider.Chat(ctx, req)
	if err != nil {
		// Check if error is retriable and try fallback once
		if b.isRetriableProviderError(err) {
			fallbackModel := b.fallback.SelectFallback(ctx, modelUsed)
			if fallbackModel != "" {
				if provider2, ok := b.registry.GetByModel(fallbackModel); ok {
					log.C(ctx).Infow("AI provider error, using fallback",
						"model", modelUsed, "fallback", fallbackModel, "err", err)
					req.Model = fallbackModel
					resp, err = provider2.Chat(ctx, req)
					if err == nil {
						// Mark quota as consumed
						quotaConsumed = true
						b.handleChatSuccess(context.Background(), uid, req.SessionID, newMessages, resp, reservedTokens)

						return resp, nil
					}
				}
			}
		}

		return nil, errno.ErrAIProviderError.WithMessage("chat failed: %v", err)
		// defer will automatically release quota
	}

	// Mark quota as consumed (will be adjusted with actual usage below)
	quotaConsumed = true
	b.handleChatSuccess(context.Background(), uid, req.SessionID, newMessages, resp, reservedTokens)

	return resp, nil
}

func (b *chatBiz) ChatStream(ctx context.Context, uid string, req *aipkg.ChatRequest) (*aipkg.ChatStream, error) {
	if len(req.Messages) == 0 {
		return nil, errno.ErrAIEmptyMessages
	}

	// Validate session if provided and get role_id from session
	if req.SessionID != "" {
		if err := b.validateSession(ctx, req.SessionID, uid); err != nil {
			return nil, err
		}

		// Get agent_id from session
		session, err := b.ds.AiSession().GetBySessionID(ctx, req.SessionID)
		if err == nil && session.AgentID != "" {
			req.AgentID = session.AgentID // Use session's agent
		}
	}

	// Apply agent preset if specified
	if err := b.buildMessagesWithAgent(ctx, req); err != nil {
		return nil, err
	}

	// Resolve model (request > session > default)
	req.Model = b.resolveModel(ctx, req.Model, req.SessionID)
	if req.Model == "" {
		return nil, errno.ErrAIModelNotFound
	}

	// Capture new messages BEFORE loading history
	newMessages := req.Messages

	// Load and merge history messages
	messages, err := b.loadAndMergeHistory(ctx, req.SessionID, req.Messages)
	if err != nil {
		return nil, err
	}
	req.Messages = messages

	// Reserve TPD quota atomically before calling provider
	reservedTokens, err := b.quota.ReserveTPD(ctx, uid, req.MaxTokens)
	if err != nil {
		return nil, err
	}

	// Ensure quota is released if not consumed (defer pattern)
	quotaConsumed := false
	defer func() {
		if !quotaConsumed && reservedTokens > 0 {
			// Release in background to avoid blocking
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
				defer cancel()
				if err := b.quota.AdjustTPD(ctx, uid, 0, reservedTokens); err != nil {
					log.C(ctx).Errorw("Failed to release reserved quota",
						"uid", uid, "reserved", reservedTokens, "err", err)
				}
			}()
		}
	}()

	// Get provider with fallback
	provider, modelUsed, err := b.getProviderWithFallback(ctx, req.Model)
	if err != nil {
		return nil, err
	}
	req.Model = modelUsed

	// Call provider
	stream, err := provider.ChatStream(ctx, req)
	if err != nil {
		// Check if error is retriable and try fallback once
		// Only attempt fallback if initial call fails (no chunks sent yet)
		if b.isRetriableProviderError(err) {
			fallbackModel := b.fallback.SelectFallback(ctx, modelUsed)
			if fallbackModel != "" {
				if provider2, ok := b.registry.GetByModel(fallbackModel); ok {
					log.C(ctx).Infow("AI provider stream error, using fallback",
						"model", modelUsed, "fallback", fallbackModel, "err", err)
					req.Model = fallbackModel
					stream, err = provider2.ChatStream(ctx, req)
					if err == nil {
						// Fallback succeeded, proceed with stream
						quotaConsumed = true

						return b.wrapStreamForSaving(stream, uid, req, newMessages, reservedTokens), nil
					}
				}
			}
		}

		return nil, errno.ErrAIProviderError.WithMessage("stream failed: %v", err)
	}

	// Mark quota as consumed (will be handled by wrapStreamForSaving)
	quotaConsumed = true

	// Wrap stream to save messages and adjust quota after completion
	return b.wrapStreamForSaving(stream, uid, req, newMessages, reservedTokens), nil
}

// wrapStreamForSaving wraps a stream to save messages and adjust quota after completion.
func (b *chatBiz) wrapStreamForSaving(stream *aipkg.ChatStream, uid string, req *aipkg.ChatRequest, newMessages []aipkg.Message, reservedTokens int) *aipkg.ChatStream {
	wrapped := aipkg.NewChatStream(aipkg.DefaultStreamBufferSize)

	go func() {
		var contentBuilder strings.Builder
		var modelName string
		var totalTokens int

		for {
			chunk, err := stream.Recv()
			if err != nil {
				// Stream ended, save accumulated content
				if contentBuilder.Len() > 0 && req.SessionID != "" {
					content := contentBuilder.String()
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
						defer cancel()
						// Pass newMessages explicitly
						b.saveStreamToSession(ctx, uid, req.SessionID, newMessages, content, modelName, totalTokens)
					}()
				}
				// Adjust TPD quota with actual usage
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
					defer cancel()
					if err := b.quota.AdjustTPD(ctx, uid, totalTokens, reservedTokens); err != nil {
						log.C(ctx).Errorw("Failed to adjust TPD quota", "uid", uid, "actual", totalTokens, "reserved", reservedTokens, "err", err)
					}
				}()
				wrapped.CloseWithError(err)

				return
			}

			// Accumulate content
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
				contentBuilder.WriteString(chunk.Choices[0].Delta.Content)
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
func (b *chatBiz) saveStreamToSession(ctx context.Context, uid string, sessionID string, newMessages []aipkg.Message, content string, modelName string, tokens int) {
	// Save user message (iterate over newMessages)
	for _, msg := range newMessages {
		if msg.Role == aipkg.RoleUser {
			if err := b.ds.AiMessage().Create(ctx, &model.AiMessageM{
				SessionID: sessionID,
				Role:      msg.Role,
				Content:   msg.Content,
				Model:     "", // User messages don't need a model
			}); err != nil {
				log.C(ctx).Errorw("Failed to save user message", "session_id", sessionID, "uid", uid, "err", err)
			}
		}
	}

	// Save assistant response
	if content != "" {
		usedModel := modelName
		if usedModel == "" {
			// Fallback if model name wasn't captured in stream
			usedModel = "unknown"
		}
		if err := b.ds.AiMessage().Create(ctx, &model.AiMessageM{
			SessionID: sessionID,
			Role:      aipkg.RoleAssistant,
			Content:   content,
			Tokens:    tokens,
			Model:     usedModel,
		}); err != nil {
			log.C(ctx).Errorw("Failed to save assistant message", "session_id", sessionID, "uid", uid, "err", err)
		}
	}

	// Update session stats
	if err := b.ds.AiSession().IncrementMessageCount(ctx, sessionID, tokens); err != nil {
		log.C(ctx).Errorw("Failed to update session stats", "session_id", sessionID, "uid", uid, "err", err)
	}
}

func (b *chatBiz) Sessions() SessionBiz {
	return NewSession(b.ds)
}

// validateSession checks if session exists and belongs to the user.
func (b *chatBiz) validateSession(ctx context.Context, sessionID, uid string) error {
	session, err := b.ds.AiSession().GetBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrAISessionNotFound
		}

		return errno.ErrOperationFailed.WithMessage("failed to get session: %v", err)
	}

	if session.UID != uid {
		return errno.ErrAISessionNotFound // Don't reveal session exists for other user
	}

	return nil
}

// resolveModel resolves the model to use based on priority:
// Request specified > Session preference > Database default > Config default > First available
func (b *chatBiz) resolveModel(ctx context.Context, reqModel, sessionID string) string {
	// 1. Request specified
	if reqModel != "" {
		return reqModel
	}

	// 2. Session preference (if session has a model set)
	if sessionID != "" {
		session, err := b.ds.AiSession().GetBySessionID(ctx, sessionID)
		if err == nil && session.Model != "" {
			return session.Model
		}
	}

	// 3. Database default model (is_default=true)
	defaultModel, err := b.ds.AiModel().GetDefault(ctx)
	if err == nil && defaultModel != nil {
		return defaultModel.Model
	}

	// 4. Config file fallback
	if facade.Config.AI.DefaultModel != "" {
		return facade.Config.AI.DefaultModel
	}

	// 5. First available model by sort order
	models, err := b.ds.AiModel().ListActive(ctx)
	if err == nil && len(models) > 0 {
		return models[0].Model
	}

	return "" // Empty string - let caller handle error
}

// getProviderWithFallback gets provider with fallback support.
// Returns (provider, actualModelUsed, error).
func (b *chatBiz) getProviderWithFallback(ctx context.Context, model string) (aipkg.Provider, string, error) {
	// First attempt: get provider for requested model
	provider, ok := b.registry.GetByModel(model)
	if ok {
		return provider, model, nil
	}

	// Fallback attempt: model not registered, try fallback
	fallbackModel := b.fallback.SelectFallback(ctx, model)
	if fallbackModel == "" {
		return nil, "", errno.ErrAIModelNotFound
	}

	provider, ok = b.registry.GetByModel(fallbackModel)
	if !ok {
		return nil, "", errno.ErrAIAllModelsFailed
	}

	return provider, fallbackModel, nil
}

// isRetriableProviderError checks if error should trigger fallback retry.
func (b *chatBiz) isRetriableProviderError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	retriable := []string{"429", "503", "502", "504", "timeout", "overloaded"}
	for _, r := range retriable {
		if strings.Contains(errMsg, r) {
			return true
		}
	}

	return false
}

// loadAndMergeHistory loads session history and merges with new messages.
// Returns merged messages with sliding window applied.
func (b *chatBiz) loadAndMergeHistory(ctx context.Context, sessionID string, newMessages []aipkg.Message) ([]aipkg.Message, error) {
	if sessionID == "" {
		return newMessages, nil
	}

	// Get limit from config
	limit := facade.Config.AI.Session.MaxMessages
	if limit <= 0 {
		limit = 50 // default
	}

	// Load history messages
	history, err := b.ds.AiMessage().ListBySessionID(ctx, sessionID, limit)
	if err != nil {
		log.C(ctx).Warnw("Failed to load message history", "session_id", sessionID, "err", err)

		return newMessages, nil // Continue without history on error
	}

	if len(history) == 0 {
		return newMessages, nil
	}

	// Convert DB messages to ai.Message
	messages := make([]aipkg.Message, 0, len(history)+len(newMessages))
	for _, m := range history {
		messages = append(messages, aipkg.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	// FIX: When we have history, only append the NEW user message(s)
	// Find the last user message in newMessages
	if len(newMessages) > 0 {
		lastUserMsgIdx := -1
		for i := len(newMessages) - 1; i >= 0; i-- {
			if newMessages[i].Role == aipkg.RoleUser {
				lastUserMsgIdx = i

				break
			}
		}

		if lastUserMsgIdx >= 0 {
			// Only include messages from the last user message onwards
			messages = append(messages, newMessages[lastUserMsgIdx:]...)
		} else {
			// No user message in newMessages, append as-is (edge case)
			messages = append(messages, newMessages...)
		}
	}

	// Apply sliding window if configured
	contextWindow := facade.Config.AI.Session.ContextWindow
	if contextWindow > 0 && len(messages) > contextWindow {
		// Keep system message if present, then last N-1 messages
		var result []aipkg.Message
		if len(messages) > 0 && messages[0].Role == aipkg.RoleSystem {
			result = append(result, messages[0])
			messages = messages[1:]
			contextWindow--
		}
		if len(messages) > contextWindow {
			messages = messages[len(messages)-contextWindow:]
		}
		result = append(result, messages...)
		messages = result
	}

	return messages, nil
}

func (b *chatBiz) ListModels(ctx context.Context) (*v1.ListModelsResponse, error) {
	models := b.registry.ListModels()

	data := make([]v1.ModelInfo, len(models))
	for i, m := range models {
		data[i] = v1.ModelInfo{
			ID:          m.ID,
			Object:      "model",
			Created:     time.Now().Unix(),
			OwnedBy:     m.Provider,
			MaxTokens:   m.MaxTokens,
			InputPrice:  m.InputPrice,
			OutputPrice: m.OutputPrice,
		}
	}

	return &v1.ListModelsResponse{
		Object: "list",
		Data:   data,
	}, nil
}

// saveToSession saves request and response to session (background goroutine)
func (b *chatBiz) saveToSession(ctx context.Context, uid string, sessionID string, newMessages []aipkg.Message, resp *aipkg.ChatResponse) {
	// Save user message (only the new ones passed in)
	for _, msg := range newMessages {
		if msg.Role == aipkg.RoleUser {
			if err := b.ds.AiMessage().Create(ctx, &model.AiMessageM{
				SessionID: sessionID,
				Role:      msg.Role,
				Content:   msg.Content,
				Model:     "", // User messages don't need a model
			}); err != nil {
				log.C(ctx).Errorw("Failed to save user message", "session_id", sessionID, "uid", uid, "err", err)
			}
		}
	}

	// Save assistant response
	if len(resp.Choices) > 0 {
		if err := b.ds.AiMessage().Create(ctx, &model.AiMessageM{
			SessionID: sessionID,
			Role:      aipkg.RoleAssistant,
			Content:   resp.Choices[0].Message.Content,
			Tokens:    resp.Usage.CompletionTokens,
			Model:     resp.Model,
		}); err != nil {
			log.C(ctx).Errorw("Failed to save assistant message", "session_id", sessionID, "uid", uid, "err", err)
		}
	}

	// Update session stats
	if err := b.ds.AiSession().IncrementMessageCount(ctx, sessionID, resp.Usage.TotalTokens); err != nil {
		log.C(ctx).Errorw("Failed to update session stats", "session_id", sessionID, "uid", uid, "err", err)
	}
}

// handleChatSuccess handles post-success operations for Chat: quota adjustment and session save.
// Called by both primary success path and fallback success path.
func (b *chatBiz) handleChatSuccess(ctx context.Context, uid string, sessionID string, newMessages []aipkg.Message, resp *aipkg.ChatResponse, reservedTokens int) {
	// Adjust TPD quota with actual usage (background)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
		defer cancel()
		if err := b.quota.AdjustTPD(ctx, uid, resp.Usage.TotalTokens, reservedTokens); err != nil {
			log.C(ctx).Errorw("Failed to adjust TPD quota",
				"uid", uid, "actual", resp.Usage.TotalTokens,
				"reserved", reservedTokens, "err", err)
		}
	}()

	// Save to session if session ID provided (background with timeout)
	if sessionID != "" {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), saveSessionTimeout)
			defer cancel()
			b.saveToSession(ctx, uid, sessionID, newMessages, resp)
		}()
	}
}

// buildMessagesWithAgent injects system prompt from agent preset if AgentID is provided.
func (b *chatBiz) buildMessagesWithAgent(ctx context.Context, req *aipkg.ChatRequest) error {
	if req.AgentID == "" {
		return nil
	}

	// Get agent details
	agent, err := b.ds.AiAgents().GetByAgentID(ctx, req.AgentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrAIRoleNotFound
		}

		return errno.ErrDBRead.WithMessage("get ai agent: %v", err)
	}

	if agent.Status == model.AiAgentStatusDisabled {
		return errno.ErrAIRoleDisabled
	}

	// Use agent model if request model is not specified or default
	if req.Model == "" || req.Model == facade.Config.AI.DefaultModel {
		if agent.Model != "" {
			req.Model = agent.Model
		}
	}

	// Apply temperature/max_tokens from agent if not custom set
	if req.Temperature == 0 && agent.Temperature > 0 {
		req.Temperature = agent.Temperature
	}
	if req.MaxTokens == 0 && agent.MaxTokens > 0 {
		req.MaxTokens = agent.MaxTokens
	}

	// Inject system prompt at the beginning
	// Check if already has system prompt to avoid duplication
	hasSystem := false
	if len(req.Messages) > 0 && req.Messages[0].Role == aipkg.RoleSystem {
		hasSystem = true
	}

	if !hasSystem && agent.SystemPrompt != "" {
		systemMsg := aipkg.Message{
			Role:    aipkg.RoleSystem,
			Content: agent.SystemPrompt,
		}
		// Prepend system message
		req.Messages = append([]aipkg.Message{systemMsg}, req.Messages...)
	}

	return nil
}
