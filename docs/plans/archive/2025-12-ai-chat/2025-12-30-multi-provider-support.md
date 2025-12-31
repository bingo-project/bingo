# Multi-Provider AI Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add Claude, Gemini, Qwen, GLM providers and seed default provider data.

**Architecture:** Each non-OpenAI-compatible provider (Claude, Gemini) gets its own implementation under `pkg/ai/providers/`. Qwen uses Eino's qwen package. GLM is OpenAI-compatible, added as preset config. All providers implement `ai.Provider` interface.

**Tech Stack:** Eino framework (eino-ext/components/model/claude, gemini, qwen), existing ai.Provider interface

---

## Task 1: Add GLM Config to OpenAI Provider

**Files:**
- Modify: `pkg/ai/providers/openai/config.go`

**Step 1: Add GLMConfig function**

```go
// GLMConfig returns configuration for Zhipu GLM (OpenAI-compatible)
func GLMConfig() *Config {
	return &Config{
		Name:    "glm",
		BaseURL: "https://open.bigmodel.cn/api/paas/v4",
		Models: []ai.ModelInfo{
			{ID: "glm-4-plus", Name: "GLM-4 Plus", Provider: "glm", MaxTokens: 128000},
			{ID: "glm-4-air", Name: "GLM-4 Air", Provider: "glm", MaxTokens: 128000},
			{ID: "glm-4-airx", Name: "GLM-4 AirX", Provider: "glm", MaxTokens: 8000},
			{ID: "glm-4-flash", Name: "GLM-4 Flash", Provider: "glm", MaxTokens: 128000},
		},
	}
}
```

**Step 2: Verify build**

Run: `make build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add pkg/ai/providers/openai/config.go
git commit -m "feat(ai): add GLM preset config for OpenAI-compatible provider"
```

---

## Task 2: Create Claude Provider

**Files:**
- Create: `pkg/ai/providers/claude/config.go`
- Create: `pkg/ai/providers/claude/provider.go`

**Step 1: Create config.go**

```go
// ABOUTME: Claude provider configuration.
// ABOUTME: Defines Config for API key and model settings.

package claude

import "github.com/bingo-project/bingo/pkg/ai"

// Config holds Claude provider configuration
type Config struct {
	APIKey string
	Models []ai.ModelInfo
}

// DefaultConfig returns default configuration for Claude
func DefaultConfig() *Config {
	return &Config{
		Models: []ai.ModelInfo{
			{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", Provider: "claude", MaxTokens: 200000},
			{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", Provider: "claude", MaxTokens: 200000},
			{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", Provider: "claude", MaxTokens: 200000},
			{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", Provider: "claude", MaxTokens: 200000},
		},
	}
}
```

**Step 2: Create provider.go**

```go
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
```

**Step 3: Add dependency**

Run: `go get github.com/cloudwego/eino-ext/components/model/claude@latest`

**Step 4: Verify build**

Run: `make build`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add pkg/ai/providers/claude/
git commit -m "feat(ai): add Claude provider implementation"
```

---

## Task 3: Create Gemini Provider

**Files:**
- Create: `pkg/ai/providers/gemini/config.go`
- Create: `pkg/ai/providers/gemini/provider.go`

**Step 1: Create config.go**

```go
// ABOUTME: Gemini provider configuration.
// ABOUTME: Defines Config for API key and model settings.

package gemini

import "github.com/bingo-project/bingo/pkg/ai"

// Config holds Gemini provider configuration
type Config struct {
	APIKey string
	Models []ai.ModelInfo
}

// DefaultConfig returns default configuration for Gemini
func DefaultConfig() *Config {
	return &Config{
		Models: []ai.ModelInfo{
			{ID: "gemini-2.0-flash-exp", Name: "Gemini 2.0 Flash", Provider: "gemini", MaxTokens: 1048576},
			{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", Provider: "gemini", MaxTokens: 2097152},
			{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", Provider: "gemini", MaxTokens: 1048576},
		},
	}
}
```

**Step 2: Create provider.go**

```go
// ABOUTME: Gemini provider implementation using Eino.
// ABOUTME: Supports chat completion and streaming for Google Gemini.

package gemini

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"time"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
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
func New(ctx context.Context, cfg *Config) (*Provider, error) {
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
```

**Step 3: Add dependency**

Run: `go get github.com/cloudwego/eino-ext/components/model/gemini@latest`

**Step 4: Verify build**

Run: `make build`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add pkg/ai/providers/gemini/
git commit -m "feat(ai): add Gemini provider implementation"
```

---

## Task 4: Create Qwen Provider

**Files:**
- Create: `pkg/ai/providers/qwen/config.go`
- Create: `pkg/ai/providers/qwen/provider.go`

**Step 1: Create config.go**

```go
// ABOUTME: Qwen provider configuration.
// ABOUTME: Defines Config for API key and model settings.

package qwen

import "github.com/bingo-project/bingo/pkg/ai"

// Config holds Qwen provider configuration
type Config struct {
	APIKey  string
	BaseURL string
	Models  []ai.ModelInfo
}

// DefaultConfig returns default configuration for Qwen
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		Models: []ai.ModelInfo{
			{ID: "qwen-max", Name: "Qwen Max", Provider: "qwen", MaxTokens: 32000},
			{ID: "qwen-plus", Name: "Qwen Plus", Provider: "qwen", MaxTokens: 131072},
			{ID: "qwen-turbo", Name: "Qwen Turbo", Provider: "qwen", MaxTokens: 131072},
			{ID: "qwen-long", Name: "Qwen Long", Provider: "qwen", MaxTokens: 10000000},
		},
	}
}
```

**Step 2: Create provider.go**

```go
// ABOUTME: Qwen provider implementation using Eino.
// ABOUTME: Supports chat completion and streaming for Alibaba Qwen.

package qwen

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"time"

	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/bingo-project/bingo/pkg/ai"
)

// Provider implements ai.Provider for Qwen
type Provider struct {
	config *Config
	client *qwen.ChatModel
}

var _ ai.Provider = (*Provider)(nil)

// New creates a new Qwen provider
func New(cfg *Config) (*Provider, error) {
	client, err := qwen.NewChatModel(context.Background(), &qwen.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Models[0].ID,
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
	return "qwen"
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
```

**Step 3: Add dependency**

Run: `go get github.com/cloudwego/eino-ext/components/model/qwen@latest`

**Step 4: Verify build**

Run: `make build`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add pkg/ai/providers/qwen/
git commit -m "feat(ai): add Qwen provider implementation"
```

---

## Task 5: Update http.go to Register All Providers

**Files:**
- Modify: `internal/apiserver/http.go`

**Step 1: Add imports and update initAIRegistry**

Add imports:
```go
import (
	// ... existing imports
	"github.com/bingo-project/bingo/pkg/ai/providers/claude"
	"github.com/bingo-project/bingo/pkg/ai/providers/gemini"
	"github.com/bingo-project/bingo/pkg/ai/providers/qwen"
)
```

Update `initAIRegistry` function:

```go
func initAIRegistry() *ai.Registry {
	credentials := facade.Config.AI.Credentials
	if len(credentials) == 0 {
		return nil
	}

	registry := ai.NewRegistry()
	ctx := context.Background()

	for name, cred := range credentials {
		var provider ai.Provider
		var err error

		switch name {
		// OpenAI-compatible providers
		case "openai":
			cfg := openai.DefaultConfig()
			cfg.APIKey = cred.APIKey
			if cred.BaseURL != "" {
				cfg.BaseURL = cred.BaseURL
			}
			provider, err = openai.New(cfg)

		case "deepseek":
			cfg := openai.DeepSeekConfig()
			cfg.APIKey = cred.APIKey
			if cred.BaseURL != "" {
				cfg.BaseURL = cred.BaseURL
			}
			provider, err = openai.New(cfg)

		case "moonshot":
			cfg := openai.MoonshotConfig()
			cfg.APIKey = cred.APIKey
			if cred.BaseURL != "" {
				cfg.BaseURL = cred.BaseURL
			}
			provider, err = openai.New(cfg)

		case "glm":
			cfg := openai.GLMConfig()
			cfg.APIKey = cred.APIKey
			if cred.BaseURL != "" {
				cfg.BaseURL = cred.BaseURL
			}
			provider, err = openai.New(cfg)

		// Native providers
		case "claude":
			cfg := claude.DefaultConfig()
			cfg.APIKey = cred.APIKey
			provider, err = claude.New(cfg)

		case "gemini":
			cfg := gemini.DefaultConfig()
			cfg.APIKey = cred.APIKey
			provider, err = gemini.New(ctx, cfg)

		case "qwen":
			cfg := qwen.DefaultConfig()
			cfg.APIKey = cred.APIKey
			if cred.BaseURL != "" {
				cfg.BaseURL = cred.BaseURL
			}
			provider, err = qwen.New(cfg)

		default:
			// Unknown provider - try as OpenAI-compatible
			cfg := openai.DefaultConfig()
			cfg.Name = name
			cfg.APIKey = cred.APIKey
			if cred.BaseURL != "" {
				cfg.BaseURL = cred.BaseURL
			}
			provider, err = openai.New(cfg)
		}

		if err != nil {
			log.Errorw("Failed to initialize AI provider", "provider", name, "err", err)
			continue
		}

		registry.Register(provider)
		log.Infow("AI provider registered", "provider", name)
	}

	return registry
}
```

**Step 2: Verify build**

Run: `make build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/apiserver/http.go
git commit -m "feat(ai): register Claude, Gemini, Qwen, GLM providers"
```

---

## Task 6: Update AI Seeder with All Providers

**Files:**
- Modify: `internal/pkg/database/seeder/ai_seeder.go`

**Step 1: Update defaultProviders and defaultModels**

```go
var defaultProviders = []model.AiProviderM{
	// OpenAI-compatible
	{Name: "openai", DisplayName: "OpenAI", Status: model.AiProviderStatusActive, Models: "[]", IsDefault: true, Sort: 1},
	{Name: "deepseek", DisplayName: "DeepSeek", Status: model.AiProviderStatusActive, Models: "[]", Sort: 2},
	{Name: "moonshot", DisplayName: "Moonshot", Status: model.AiProviderStatusActive, Models: "[]", Sort: 3},
	{Name: "glm", DisplayName: "智谱 GLM", Status: model.AiProviderStatusActive, Models: "[]", Sort: 4},
	// Native providers
	{Name: "claude", DisplayName: "Claude", Status: model.AiProviderStatusActive, Models: "[]", Sort: 5},
	{Name: "gemini", DisplayName: "Gemini", Status: model.AiProviderStatusActive, Models: "[]", Sort: 6},
	{Name: "qwen", DisplayName: "通义千问", Status: model.AiProviderStatusActive, Models: "[]", Sort: 7},
}

var defaultModels = []model.AiModelM{
	// OpenAI
	{ProviderName: "openai", Model: "gpt-4o", DisplayName: "GPT-4o", MaxTokens: 128000, Status: model.AiModelStatusActive, IsDefault: true, Sort: 1},
	{ProviderName: "openai", Model: "gpt-4o-mini", DisplayName: "GPT-4o Mini", MaxTokens: 128000, Status: model.AiModelStatusActive, Sort: 2},
	{ProviderName: "openai", Model: "gpt-4-turbo", DisplayName: "GPT-4 Turbo", MaxTokens: 128000, Status: model.AiModelStatusActive, Sort: 3},
	{ProviderName: "openai", Model: "gpt-3.5-turbo", DisplayName: "GPT-3.5 Turbo", MaxTokens: 16385, Status: model.AiModelStatusActive, Sort: 4},

	// DeepSeek
	{ProviderName: "deepseek", Model: "deepseek-chat", DisplayName: "DeepSeek Chat", MaxTokens: 64000, Status: model.AiModelStatusActive, Sort: 1},
	{ProviderName: "deepseek", Model: "deepseek-coder", DisplayName: "DeepSeek Coder", MaxTokens: 64000, Status: model.AiModelStatusActive, Sort: 2},

	// Moonshot
	{ProviderName: "moonshot", Model: "moonshot-v1-8k", DisplayName: "Moonshot V1 8K", MaxTokens: 8000, Status: model.AiModelStatusActive, Sort: 1},
	{ProviderName: "moonshot", Model: "moonshot-v1-32k", DisplayName: "Moonshot V1 32K", MaxTokens: 32000, Status: model.AiModelStatusActive, Sort: 2},
	{ProviderName: "moonshot", Model: "moonshot-v1-128k", DisplayName: "Moonshot V1 128K", MaxTokens: 128000, Status: model.AiModelStatusActive, Sort: 3},

	// GLM (智谱)
	{ProviderName: "glm", Model: "glm-4-plus", DisplayName: "GLM-4 Plus", MaxTokens: 128000, Status: model.AiModelStatusActive, Sort: 1},
	{ProviderName: "glm", Model: "glm-4-air", DisplayName: "GLM-4 Air", MaxTokens: 128000, Status: model.AiModelStatusActive, Sort: 2},
	{ProviderName: "glm", Model: "glm-4-airx", DisplayName: "GLM-4 AirX", MaxTokens: 8000, Status: model.AiModelStatusActive, Sort: 3},
	{ProviderName: "glm", Model: "glm-4-flash", DisplayName: "GLM-4 Flash", MaxTokens: 128000, Status: model.AiModelStatusActive, Sort: 4},

	// Claude
	{ProviderName: "claude", Model: "claude-sonnet-4-20250514", DisplayName: "Claude Sonnet 4", MaxTokens: 200000, Status: model.AiModelStatusActive, Sort: 1},
	{ProviderName: "claude", Model: "claude-3-5-sonnet-20241022", DisplayName: "Claude 3.5 Sonnet", MaxTokens: 200000, Status: model.AiModelStatusActive, Sort: 2},
	{ProviderName: "claude", Model: "claude-3-5-haiku-20241022", DisplayName: "Claude 3.5 Haiku", MaxTokens: 200000, Status: model.AiModelStatusActive, Sort: 3},
	{ProviderName: "claude", Model: "claude-3-opus-20240229", DisplayName: "Claude 3 Opus", MaxTokens: 200000, Status: model.AiModelStatusActive, Sort: 4},

	// Gemini
	{ProviderName: "gemini", Model: "gemini-2.0-flash-exp", DisplayName: "Gemini 2.0 Flash", MaxTokens: 1048576, Status: model.AiModelStatusActive, Sort: 1},
	{ProviderName: "gemini", Model: "gemini-1.5-pro", DisplayName: "Gemini 1.5 Pro", MaxTokens: 2097152, Status: model.AiModelStatusActive, Sort: 2},
	{ProviderName: "gemini", Model: "gemini-1.5-flash", DisplayName: "Gemini 1.5 Flash", MaxTokens: 1048576, Status: model.AiModelStatusActive, Sort: 3},

	// Qwen
	{ProviderName: "qwen", Model: "qwen-max", DisplayName: "Qwen Max", MaxTokens: 32000, Status: model.AiModelStatusActive, Sort: 1},
	{ProviderName: "qwen", Model: "qwen-plus", DisplayName: "Qwen Plus", MaxTokens: 131072, Status: model.AiModelStatusActive, Sort: 2},
	{ProviderName: "qwen", Model: "qwen-turbo", DisplayName: "Qwen Turbo", MaxTokens: 131072, Status: model.AiModelStatusActive, Sort: 3},
	{ProviderName: "qwen", Model: "qwen-long", DisplayName: "Qwen Long", MaxTokens: 10000000, Status: model.AiModelStatusActive, Sort: 4},
}
```

**Step 2: Verify build**

Run: `make build`
Expected: Build succeeds

**Step 3: Run seeder to test**

Run: `bingo db seed --seeder=AiSeeder`
Expected: Seeder completes without errors

**Step 4: Commit**

```bash
git add internal/pkg/database/seeder/ai_seeder.go
git commit -m "feat(ai): seed all supported providers and models"
```

---

## Task 7: Final Verification

**Step 1: Run full build and lint**

```bash
make lint
make build
```

Expected: All checks pass

**Step 2: Create final commit if needed**

```bash
git add -A
git commit -m "chore: cleanup after multi-provider implementation"
```

---

## Summary

| Provider | Type | Eino Package | Status |
|----------|------|--------------|--------|
| OpenAI | OpenAI-compatible | eino-ext/model/openai | ✅ |
| DeepSeek | OpenAI-compatible | eino-ext/model/openai | ✅ |
| Moonshot | OpenAI-compatible | eino-ext/model/openai | ✅ |
| GLM | OpenAI-compatible | eino-ext/model/openai | ✅ |
| Claude | Native | eino-ext/model/claude | ✅ |
| Gemini | Native | eino-ext/model/gemini | ✅ |
| Qwen | Native | eino-ext/model/qwen | ✅ |

**Config example (configs/apiserver.yaml):**

```yaml
ai:
  credentials:
    openai:
      api-key: "sk-xxx"
    claude:
      api-key: "sk-ant-xxx"
    gemini:
      api-key: "xxx"
    qwen:
      api-key: "sk-xxx"
    glm:
      api-key: "xxx"
    deepseek:
      api-key: "sk-xxx"
```
