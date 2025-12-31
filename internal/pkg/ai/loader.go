// ABOUTME: AI provider loader from database configuration.
// ABOUTME: Loads active providers and models from Store into Registry.

package ai

import (
	"context"
	"fmt"

	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	aipkg "github.com/bingo-project/bingo/pkg/ai"
	"github.com/bingo-project/bingo/pkg/ai/providers/claude"
	"github.com/bingo-project/bingo/pkg/ai/providers/gemini"
	"github.com/bingo-project/bingo/pkg/ai/providers/openai"
	"github.com/bingo-project/bingo/pkg/ai/providers/qwen"
)

// Credential holds provider authentication configuration.
type Credential struct {
	APIKey  string
	BaseURL string
}

// Loader loads AI providers from database into Registry.
type Loader struct {
	registry    *aipkg.Registry
	store       store.IStore
	credentials map[string]Credential
}

// NewLoader creates a new Loader.
func NewLoader(registry *aipkg.Registry, st store.IStore, creds map[string]Credential) *Loader {
	return &Loader{
		registry:    registry,
		store:       st,
		credentials: creds,
	}
}

// Load loads active providers from database into registry.
func (l *Loader) Load(ctx context.Context) error {
	providers, err := l.store.AiProvider().ListActive(ctx)
	if err != nil {
		return fmt.Errorf("list active providers: %w", err)
	}

	models, err := l.store.AiModel().ListActive(ctx)
	if err != nil {
		return fmt.Errorf("list active models: %w", err)
	}

	modelsByProvider := l.groupModelsByProvider(models)

	for _, p := range providers {
		if err := l.loadProvider(ctx, p, modelsByProvider[p.Name]); err != nil {
			log.Errorw("Failed to load AI provider", "provider", p.Name, "err", err)
		}
	}

	return nil
}

// Reload clears and reloads all providers.
func (l *Loader) Reload(ctx context.Context) error {
	l.registry.Clear()

	return l.Load(ctx)
}

// groupModelsByProvider groups models by provider name.
func (l *Loader) groupModelsByProvider(models []*model.AiModelM) map[string][]*model.AiModelM {
	result := make(map[string][]*model.AiModelM)
	for _, m := range models {
		result[m.ProviderName] = append(result[m.ProviderName], m)
	}

	return result
}

// loadProvider creates and registers a single provider.
func (l *Loader) loadProvider(ctx context.Context, p *model.AiProviderM, models []*model.AiModelM) error {
	cred, hasCred := l.credentials[p.Name]
	if !hasCred {
		log.Warnw("AI provider has no credential, skipping", "provider", p.Name)

		return nil
	}

	provider, err := l.createProvider(p.Name, cred, models)
	if err != nil {
		return err
	}

	l.registry.Register(provider)
	log.Infow("AI provider loaded", "provider", p.Name)

	return nil
}

// createProvider instantiates a provider by name.
func (l *Loader) createProvider(name string, cred Credential, models []*model.AiModelM) (aipkg.Provider, error) {
	modelInfos := make([]aipkg.ModelInfo, len(models))
	for i, m := range models {
		modelInfos[i] = aipkg.ModelInfo{
			ID:        m.Model,
			Name:      m.DisplayName,
			Provider:  m.ProviderName,
			MaxTokens: m.MaxTokens,
		}
	}

	switch name {
	case "openai":
		cfg := openai.DefaultConfig()
		cfg.APIKey = cred.APIKey
		cfg.BaseURL = cred.BaseURL
		cfg.Models = modelInfos

		return openai.New(cfg)

	case "deepseek":
		cfg := openai.DeepSeekConfig()
		cfg.APIKey = cred.APIKey
		cfg.BaseURL = cred.BaseURL
		cfg.Models = modelInfos

		return openai.New(cfg)

	case "moonshot":
		cfg := openai.MoonshotConfig()
		cfg.APIKey = cred.APIKey
		cfg.BaseURL = cred.BaseURL
		cfg.Models = modelInfos

		return openai.New(cfg)

	case "glm":
		cfg := openai.GLMConfig()
		cfg.APIKey = cred.APIKey
		cfg.BaseURL = cred.BaseURL
		cfg.Models = modelInfos

		return openai.New(cfg)

	case "claude":
		cfg := claude.DefaultConfig()
		cfg.APIKey = cred.APIKey
		cfg.Models = modelInfos

		return claude.New(cfg)

	case "gemini":
		cfg := gemini.DefaultConfig()
		cfg.APIKey = cred.APIKey
		cfg.Models = modelInfos

		return gemini.New(cfg)

	case "qwen":
		cfg := qwen.DefaultConfig()
		cfg.APIKey = cred.APIKey
		cfg.BaseURL = cred.BaseURL
		cfg.Models = modelInfos

		return qwen.New(cfg)

	default:
		cfg := openai.DefaultConfig()
		cfg.Name = name
		cfg.APIKey = cred.APIKey
		cfg.BaseURL = cred.BaseURL
		cfg.Models = modelInfos

		return openai.New(cfg)
	}
}
