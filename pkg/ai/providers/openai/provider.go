// ABOUTME: OpenAI provider implementation using Eino.
// ABOUTME: Supports chat completion and streaming for OpenAI-compatible APIs.

package openai

import (
	"context"
	"io"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/bingo-project/bingo/pkg/ai"
)

// Provider implements ai.Provider for OpenAI
type Provider struct {
	config *Config
	client *openai.ChatModel
}

var _ ai.Provider = (*Provider)(nil)

// New creates a new OpenAI provider
func New(cfg *Config) (*Provider, error) {
	client, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Timeout: 120 * time.Second,
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
	return "openai"
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

	// Start goroutine to read from Eino stream
	go func() {
		defer chatStream.Close()

		for {
			chunk, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					chatStream.CloseWithError(err)
				}

				return
			}

			chatStream.Send(convertStreamChunk(chunk, req.Model))
		}
	}()

	return chatStream, nil
}

// convertMessages converts ai.Message to schema.Message
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

// convertResponse converts Eino response to ai.ChatResponse
func convertResponse(resp *schema.Message, modelName string) *ai.ChatResponse {
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
		Usage: ai.Usage{},
	}
}

// convertStreamChunk converts Eino stream message to ai.StreamChunk
func convertStreamChunk(msg *schema.Message, modelName string) *ai.StreamChunk {
	return &ai.StreamChunk{
		ID:      generateID(),
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
}

// generateID generates a unique ID for responses
func generateID() string {
	return "chatcmpl-" + time.Now().Format("20060102150405")
}
