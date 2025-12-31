// ABOUTME: Gemini provider implementation using Eino.
// ABOUTME: Supports chat completion and streaming for Google Gemini.

package gemini

import (
	"context"
	"io"
	"time"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino/components/model"
	"google.golang.org/genai"

	"github.com/bingo-project/bingo/pkg/ai"
)

// Provider implements ai.Provider for Gemini
type Provider struct {
	config *Config
	client *gemini.ChatModel
}

var _ ai.Provider = (*Provider)(nil)

// New creates a new Gemini provider
func New(cfg *Config) (*Provider, error) {
	ctx := context.Background()
	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	client, err := gemini.NewChatModel(ctx, &gemini.Config{
		Client: genaiClient,
		Model:  cfg.Models[0].ID,
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
	return "gemini"
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

				chatStream.Send(ai.ConvertStreamChunk(chunk, req.Model, id))
			}
		}
	}()

	return chatStream, nil
}
