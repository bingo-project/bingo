// ABOUTME: OpenAI provider implementation using Eino.
// ABOUTME: Supports chat completion and streaming for OpenAI-compatible APIs.

package openai

import (
	"context"
	"io"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"

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
	if p.config.Name != "" {
		return p.config.Name
	}

	return "openai"
}

// Models returns available models
func (p *Provider) Models() []ai.ModelInfo {
	return p.config.Models
}

// Chat performs a non-streaming chat completion
func (p *Provider) Chat(ctx context.Context, req *ai.ChatRequest) (*ai.ChatResponse, error) {
	messages := ai.ConvertMessages(req.Messages)

	opts := []model.Option{
		model.WithModel(req.Model),
	}
	if req.MaxTokens > 0 {
		opts = append(opts, model.WithMaxTokens(req.MaxTokens))
	}
	if req.Temperature > 0 {
		opts = append(opts, model.WithTemperature(float32(req.Temperature)))
	}

	return ai.Do(ctx, ai.DefaultRetryConfig, func(ctx context.Context) (*ai.ChatResponse, error) {
		resp, err := p.client.Generate(ctx, messages, opts...)
		if err != nil {
			return nil, err
		}

		return ai.ConvertResponse(resp, req.Model), nil
	})
}

// ChatStream performs a streaming chat completion
func (p *Provider) ChatStream(ctx context.Context, req *ai.ChatRequest) (*ai.ChatStream, error) {
	messages := ai.ConvertMessages(req.Messages)

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

	chatStream := ai.NewChatStream(ai.DefaultStreamBufferSize)

	// Start goroutine to read from Eino stream
	go func() {
		defer chatStream.Close()

		id := ai.GenerateID()
		var lastUsage *ai.Usage

		for {
			select {
			case <-ctx.Done():
				chatStream.CloseWithError(ctx.Err())
				return
			default:
				chunk, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						// Send final chunk with finish_reason and usage
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

				// Track usage from chunks (Eino sends it in the last content chunk)
				if chunk.ResponseMeta != nil && chunk.ResponseMeta.Usage != nil {
					lastUsage = &ai.Usage{
						PromptTokens:     chunk.ResponseMeta.Usage.PromptTokens,
						CompletionTokens: chunk.ResponseMeta.Usage.CompletionTokens,
						TotalTokens:      chunk.ResponseMeta.Usage.TotalTokens,
					}
				}

				chatStream.Send(ai.ConvertStreamChunk(chunk, req.Model, id))
			}
		}
	}()

	return chatStream, nil
}
