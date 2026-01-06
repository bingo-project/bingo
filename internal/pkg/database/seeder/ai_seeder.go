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
	{Name: "openai", DisplayName: "OpenAI", Status: model.AiProviderStatusActive, IsDefault: true, Sort: 1},
	{Name: "deepseek", DisplayName: "DeepSeek", Status: model.AiProviderStatusActive, Sort: 2},
	{Name: "moonshot", DisplayName: "Moonshot", Status: model.AiProviderStatusActive, Sort: 3},
	{Name: "glm", DisplayName: "智谱 GLM", Status: model.AiProviderStatusActive, Sort: 4},
	// Native providers
	{Name: "claude", DisplayName: "Claude", Status: model.AiProviderStatusActive, Sort: 5},
	{Name: "gemini", DisplayName: "Gemini", Status: model.AiProviderStatusActive, Sort: 6},
	{Name: "qwen", DisplayName: "通义千问", Status: model.AiProviderStatusActive, Sort: 7},
}

var defaultModels = []model.AiModelM{
	// OpenAI
	{ProviderName: "openai", Model: "gpt-4o", DisplayName: "GPT-4o", MaxTokens: 128000, Status: model.AiModelStatusDisabled, Sort: 1},
	{ProviderName: "openai", Model: "gpt-4o-mini", DisplayName: "GPT-4o Mini", MaxTokens: 128000, Status: model.AiModelStatusDisabled, Sort: 2},
	{ProviderName: "openai", Model: "gpt-4-turbo", DisplayName: "GPT-4 Turbo", MaxTokens: 128000, Status: model.AiModelStatusDisabled, Sort: 3},
	{ProviderName: "openai", Model: "gpt-3.5-turbo", DisplayName: "GPT-3.5 Turbo", MaxTokens: 16385, Status: model.AiModelStatusDisabled, Sort: 4},

	// DeepSeek
	{ProviderName: "deepseek", Model: "deepseek-chat", DisplayName: "DeepSeek Chat", MaxTokens: 64000, Status: model.AiModelStatusDisabled, Sort: 1},
	{ProviderName: "deepseek", Model: "deepseek-coder", DisplayName: "DeepSeek Coder", MaxTokens: 64000, Status: model.AiModelStatusDisabled, Sort: 2},

	// Moonshot
	{ProviderName: "moonshot", Model: "moonshot-v1-8k", DisplayName: "Moonshot V1 8K", MaxTokens: 8000, Status: model.AiModelStatusDisabled, Sort: 1},
	{ProviderName: "moonshot", Model: "moonshot-v1-32k", DisplayName: "Moonshot V1 32K", MaxTokens: 32000, Status: model.AiModelStatusDisabled, Sort: 2},
	{ProviderName: "moonshot", Model: "moonshot-v1-128k", DisplayName: "Moonshot V1 128K", MaxTokens: 128000, Status: model.AiModelStatusDisabled, Sort: 3},

	// GLM (智谱) - Active models
	{ProviderName: "glm", Model: "glm-4.5-plus", DisplayName: "GLM-4.5 Plus", MaxTokens: 128000, Status: model.AiModelStatusActive, Sort: 1},
	{ProviderName: "glm", Model: "glm-4.5-air", DisplayName: "GLM-4.5 Air", MaxTokens: 128000, Status: model.AiModelStatusActive, Sort: 2},
	{ProviderName: "glm", Model: "glm-4.5-airx", DisplayName: "GLM-4.5 AirX", MaxTokens: 8000, Status: model.AiModelStatusActive, Sort: 3},
	{ProviderName: "glm", Model: "glm-4.5-flash", DisplayName: "GLM-4.5 Flash", MaxTokens: 128000, Status: model.AiModelStatusActive, IsDefault: true, Sort: 4},
	{ProviderName: "glm", Model: "glm-4.7", DisplayName: "GLM-4.7", MaxTokens: 128000, Status: model.AiModelStatusActive, Sort: 5},

	// Claude
	{ProviderName: "claude", Model: "claude-sonnet-4-20250514", DisplayName: "Claude Sonnet 4", MaxTokens: 200000, Status: model.AiModelStatusDisabled, Sort: 1},
	{ProviderName: "claude", Model: "claude-3-5-sonnet-20241022", DisplayName: "Claude 3.5 Sonnet", MaxTokens: 200000, Status: model.AiModelStatusDisabled, Sort: 2},
	{ProviderName: "claude", Model: "claude-3-5-haiku-20241022", DisplayName: "Claude 3.5 Haiku", MaxTokens: 200000, Status: model.AiModelStatusDisabled, Sort: 3},
	{ProviderName: "claude", Model: "claude-3-opus-20240229", DisplayName: "Claude 3 Opus", MaxTokens: 200000, Status: model.AiModelStatusDisabled, Sort: 4},

	// Gemini
	{ProviderName: "gemini", Model: "gemini-2.0-flash-exp", DisplayName: "Gemini 2.0 Flash", MaxTokens: 1048576, Status: model.AiModelStatusDisabled, Sort: 1},
	{ProviderName: "gemini", Model: "gemini-1.5-pro", DisplayName: "Gemini 1.5 Pro", MaxTokens: 2097152, Status: model.AiModelStatusDisabled, Sort: 2},
	{ProviderName: "gemini", Model: "gemini-1.5-flash", DisplayName: "Gemini 1.5 Flash", MaxTokens: 1048576, Status: model.AiModelStatusDisabled, Sort: 3},

	// Qwen
	{ProviderName: "qwen", Model: "qwen-max", DisplayName: "Qwen Max", MaxTokens: 32000, Status: model.AiModelStatusDisabled, Sort: 1},
	{ProviderName: "qwen", Model: "qwen-plus", DisplayName: "Qwen Plus", MaxTokens: 131072, Status: model.AiModelStatusDisabled, Sort: 2},
	{ProviderName: "qwen", Model: "qwen-turbo", DisplayName: "Qwen Turbo", MaxTokens: 131072, Status: model.AiModelStatusDisabled, Sort: 3},
	{ProviderName: "qwen", Model: "qwen-long", DisplayName: "Qwen Long", MaxTokens: 10000000, Status: model.AiModelStatusDisabled, Sort: 4},
}

var defaultAiAgents = []model.AiAgentM{
	{
		AgentID:     "coding_expert",
		Name:        "全栈代码专家",
		Description: "精通 Go/Python/Vue 的技术专家，代码优先，注释详细",
		Model:       "glm-4.5-flash",
		Category:    model.AiAgentCategoryWorkplace,
		Status:      model.AiAgentStatusActive,
		Sort:        1,
		SystemPrompt: `# 角色设定
你是一位拥有 10 年经验的全栈技术专家，精通 Golang (CloudWeGo/Gin)、Python、Vue3 和系统架构设计。你的代码风格简洁、高效且符合工程最佳实践。

# 回复规范
1. **代码优先**：直接给出解决方案的代码，代码块必须指定语言（如 ` + "`" + `go）。
2. **详细注释**：关键逻辑必须包含中文注释，解释"为什么这么写"。
3. **原理解析**：代码之后，简要解释实现原理和潜在的坑（Edge Cases）。
4. **拒绝废话**：不要说"希望这对你有帮助"之类的客套话。`,
	},
	{
		AgentID:     "translator_pro",
		Name:        "多语言翻译官",
		Description: "沉浸式翻译体验，自动识别中英互译",
		Model:       "glm-4.5-flash",
		Category:    model.AiAgentCategoryGeneral,
		Status:      model.AiAgentStatusActive,
		Sort:        2,
		SystemPrompt: `# 角色设定
你是一位精通多国语言的专业翻译官，致力于提供"信、达、雅"的翻译服务。

# 翻译规则
1. **直接输出**：不要输出任何解释性文字，只输出翻译结果。
2. **智能识别**：自动识别输入语言。如果是中文，则翻译成英文；如果是英文，则翻译成中文。
3. **风格化**：
   - 默认风格：商务专业，适合邮件和文档。
   - 如果用户指定"口语化"，则使用更自然的日常表达。`,
	},
	{
		AgentID:     "tech_writer",
		Name:        "技术文档专家",
		Description: "编写清晰、结构化且美观的 Markdown 技术文档",
		Model:       "glm-4.5-flash",
		Category:    model.AiAgentCategoryWorkplace,
		Status:      model.AiAgentStatusActive,
		Sort:        3,
		SystemPrompt: `# 角色设定
你是一位专业的技术文档工程师，擅长编写清晰、结构化且易于阅读的技术文档。

# 写作规范
1. **结构清晰**：使用正确的 Markdown 标题层级 (#, ##, ###)。
2. **排版美观**：适当使用列表、引用和加粗来强调重点。
3. **Emoji 装饰**：在标题或关键点前使用适当的 Emoji 图标，增加可读性。
4. **专业术语**：确保技术术语准确，必要时提供英文对照。
5. **自动目录**：长文档请在开头提供目录 (TOC)。`,
	},
	{
		AgentID:     "prompt_optimizer",
		Name:        "提示词优化师",
		Description: "将模糊需求转化为高质量的结构化 System Prompt",
		Model:       "glm-4.5-flash",
		Category:    model.AiAgentCategoryCreative,
		Status:      model.AiAgentStatusActive,
		Sort:        4,
		SystemPrompt: `# 角色设定
你是一位资深的 Prompt 工程师，擅长将用户的模糊需求转化为结构化、高质量的 System Prompt。

# 任务目标
根据用户输入的需求，输出一个优化后的 System Prompt。

# 输出格式
请使用 Markdown 代码块输出优化后的 Prompt，结构应包含：
1. **角色设定** (Role)
2. **任务目标** (Objective)
3. **约束条件** (Constraints)
4. **输出格式** (Workflow/Output Format)
5. **示例** (Examples, 可选)

# 示例
用户输入："帮我写个改写文章的 prompt"
你回复：(一个完整的 Prompt 结构代码块)`,
	},
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
		where := &model.AiModelM{ProviderName: m.ProviderName, Model: m.Model}
		if err := store.S.AiModel().FirstOrCreate(ctx, where, &m); err != nil {
			return err
		}
	}

	// Seed agents
	for _, agent := range defaultAiAgents {
		where := &model.AiAgentM{AgentID: agent.AgentID}
		// Because SystemPrompt is long text, FirstOrCreate might not update it if the record exists but has old content.
		// For seeding, we generally want to enforce the latest content for built-in agents.
		// However, standard FirstOrCreate behavior is "do nothing if exists".
		// We trust FirstOrCreate for now. If we need to force update, we'd check and update.
		if err := store.S.AiAgents().FirstOrCreate(ctx, where, &agent); err != nil {
			return err
		}
	}

	return nil
}
