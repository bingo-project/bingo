// ABOUTME: Seeds default AI configuration data.
// ABOUTME: Initializes quota tiers, providers, and models.

package seeder

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

var defaultQuotaTiers = []model.AiQuotaTierM{
	{Tier: model.AiQuotaTierFree, DisplayName: "Free Tier", RPM: 10, TPD: 100000},
	{Tier: model.AiQuotaTierPro, DisplayName: "Pro Tier", RPM: 60, TPD: 1000000},
	{Tier: model.AiQuotaTierEnterprise, DisplayName: "Enterprise Tier", RPM: 300, TPD: 10000000},
}

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

type AiSeeder struct{}

func (AiSeeder) Signature() string {
	return "AiSeeder"
}

func (AiSeeder) Run() error {
	ctx := context.Background()

	// Seed quota tiers
	for _, tier := range defaultQuotaTiers {
		where := &model.AiQuotaTierM{Tier: tier.Tier}
		if err := store.S.AiQuotaTier().FirstOrCreate(ctx, where, &tier); err != nil {
			return err
		}
	}

	// Seed providers
	for _, provider := range defaultProviders {
		where := &model.AiProviderM{Name: provider.Name}
		if err := store.S.AiProvider().FirstOrCreate(ctx, where, &provider); err != nil {
			return err
		}
	}

	// Seed models
	for _, m := range defaultModels {
		where := &model.AiModelM{Model: m.Model}
		if err := store.S.AiModel().FirstOrCreate(ctx, where, &m); err != nil {
			return err
		}
	}

	return nil
}
