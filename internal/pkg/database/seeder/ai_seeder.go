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
	{Name: "openai", DisplayName: "OpenAI", Status: model.AiProviderStatusActive, Models: "[]", IsDefault: true, Sort: 1},
}

var defaultModels = []model.AiModelM{
	{ProviderName: "openai", Model: "gpt-4o", DisplayName: "GPT-4o", MaxTokens: 128000, Status: model.AiModelStatusActive, IsDefault: true, Sort: 1},
	{ProviderName: "openai", Model: "gpt-4o-mini", DisplayName: "GPT-4o Mini", MaxTokens: 128000, Status: model.AiModelStatusActive, Sort: 2},
	{ProviderName: "openai", Model: "gpt-4-turbo", DisplayName: "GPT-4 Turbo", MaxTokens: 128000, Status: model.AiModelStatusActive, Sort: 3},
	{ProviderName: "openai", Model: "gpt-3.5-turbo", DisplayName: "GPT-3.5 Turbo", MaxTokens: 16385, Status: model.AiModelStatusActive, Sort: 4},
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
