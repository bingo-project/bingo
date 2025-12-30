// ABOUTME: Claude provider implementation using Eino.
// ABOUTME: Supports chat completion and streaming for Anthropic Claude.

package claude

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"time"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/bingo-project/bingo/pkg/ai"
)

// Provider implements ai.Provider for Claude
type Provider struct {
	config *Config
	client *claude.ChatModel
}

var _ ai.Provider = (*Provider)(nil)

// New creates a new Claude provider
func New(cfg *Config) (*Provider, error) {
	client, err := claude.NewChatModel(context.Background(), &claude.Config{
		APIKey:    cfg.APIKey,
		Model:     cfg.Models[0].ID, // Default model
		MaxTokens: 4096,
	})
	if err != nil {
		return nil, err
	}

	return &Provider{
		config: cfg,
		client: client,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "claude"
}

// Models returns available models
func (p *Provider) Models() []ai.ModelInfo {
	return p.config.Models
}

// Chat performs a non-streaming chat completion
func (p *Provider) Chat(ctx context.Context, req *ai.ChatRequest) (*ai.ChatResponse, error) {
	messages := convertMessages(req.Messages)

	opts := []model.Option{
		model.WithModel(req.Model),
	}
	if req.MaxTokens > 0 {
		opts = append(opts, model.WithMaxTokens(req.MaxTokens))
	}
	if req.Temperature > 0 {
		opts = append(opts, model.WithTemperature(float32(req.Temperature)))
	}

	resp, err := p.client.Generate(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}

	return convertResponse(resp, req.Model), nil
}

// ChatStream performs a streaming chat completion
func (p *Provider) ChatStream(ctx context.Context, req *ai.ChatRequest) (*ai.ChatStream, error) {
	messages := convertMessages(req.Messages)

	opts := []model.Option{
		model.WithModel(req.Model),
	}
	if req.MaxTokens > 0 {
		opts = append(opts, model.WithMaxTokens(req.MaxTokens))
	}
	if req.Temperature > 0 {
		opts = append(opts, model.WithTemperature(float32(req.Temperature)))
	}

	stream, err := p.client.Stream(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}

	chatStream := ai.NewChatStream(100)

	go func() {
		defer chatStream.Close()

		id := generateID()
		var lastUsage *ai.Usage

		for {
			chunk, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					chatStream.Send(&ai.StreamChunk{
						ID:      id,
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   req.Model,
						Choices: []ai.Choice{
							{
								Index:        0,
								Delta:        &ai.Message{},
								FinishReason: "stop",
							},
						},
						Usage: lastUsage,
					})
				} else {
					chatStream.CloseWithError(err)
				}

				return
			}

			if chunk.ResponseMeta != nil && chunk.ResponseMeta.Usage != nil {
				lastUsage = &ai.Usage{
					PromptTokens:     chunk.ResponseMeta.Usage.PromptTokens,
					CompletionTokens: chunk.ResponseMeta.Usage.CompletionTokens,
					TotalTokens:      chunk.ResponseMeta.Usage.TotalTokens,
				}
			}

			chatStream.Send(convertStreamChunk(chunk, req.Model, id))
		}
	}()

	return chatStream, nil
}

func convertMessages(msgs []ai.Message) []*schema.Message {
	result := make([]*schema.Message, len(msgs))
	for i, m := range msgs {
		role := schema.User
		switch m.Role {
		case ai.RoleSystem:
			role = schema.System
		case ai.RoleAssistant:
			role = schema.Assistant
		}
		result[i] = &schema.Message{
			Role:    role,
			Content: m.Content,
		}
	}

	return result
}

func convertResponse(resp *schema.Message, modelName string) *ai.ChatResponse {
	usage := extractUsage(resp)

	return &ai.ChatResponse{
		ID:      generateID(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   modelName,
		Choices: []ai.Choice{
			{
				Index: 0,
				Message: ai.Message{
					Role:    ai.RoleAssistant,
					Content: resp.Content,
				},
				FinishReason: "stop",
			},
		},
		Usage: usage,
	}
}

func convertStreamChunk(msg *schema.Message, modelName string, id string) *ai.StreamChunk {
	chunk := &ai.StreamChunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   modelName,
		Choices: []ai.Choice{
			{
				Index: 0,
				Delta: &ai.Message{
					Role:    ai.RoleAssistant,
					Content: msg.Content,
				},
			},
		},
	}

	if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
		chunk.Usage = &ai.Usage{
			PromptTokens:     msg.ResponseMeta.Usage.PromptTokens,
			CompletionTokens: msg.ResponseMeta.Usage.CompletionTokens,
			TotalTokens:      msg.ResponseMeta.Usage.TotalTokens,
		}
	}

	return chunk
}

func generateID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)

	return "chatcmpl-" + hex.EncodeToString(b)
}

func extractUsage(msg *schema.Message) ai.Usage {
	if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
		return ai.Usage{
			PromptTokens:     msg.ResponseMeta.Usage.PromptTokens,
			CompletionTokens: msg.ResponseMeta.Usage.CompletionTokens,
			TotalTokens:      msg.ResponseMeta.Usage.TotalTokens,
		}
	}

	return ai.Usage{}
}
