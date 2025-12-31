// ABOUTME: HTTP server initialization for apiserver.
// ABOUTME: Configures Gin engine with routes and middleware.

package apiserver

import (
	"github.com/gin-gonic/gin"

	bizauth "github.com/bingo-project/bingo/internal/apiserver/biz/auth"
	"github.com/bingo-project/bingo/internal/apiserver/middleware"
	"github.com/bingo-project/bingo/internal/apiserver/router"
	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/bootstrap"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/log"
	httpmw "github.com/bingo-project/bingo/internal/pkg/middleware/http"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ai"
	"github.com/bingo-project/bingo/pkg/ai/providers/claude"
	"github.com/bingo-project/bingo/pkg/ai/providers/gemini"
	"github.com/bingo-project/bingo/pkg/ai/providers/openai"
	"github.com/bingo-project/bingo/pkg/ai/providers/qwen"
)

// initGinEngine initializes the Gin engine with routes.
func initGinEngine() *gin.Engine {
	g := bootstrap.InitGin()

	// Swagger
	if facade.Config.Feature.ApiDoc {
		router.MapSwagRouters(g)
	}

	// Common router
	router.MapCommonRouters(g)

	// Api
	router.MapApiRouters(g)

	// AI Chat routes
	if registry := initAIRegistry(); registry != nil {
		v1 := g.Group("/v1")
		v1.Use(middleware.Lang())
		v1.Use(middleware.Maintenance())

		loader := bizauth.NewUserLoader(store.S)
		authn := auth.New(loader)
		v1.Use(auth.Middleware(authn))

		// Apply AI rate limiter (RPM)
		rpm := facade.Config.AI.Quota.DefaultRPM
		if rpm <= 0 {
			rpm = 20 // fallback default
		}
		v1.Use(httpmw.AILimiter(rpm))

		router.MapChatRouters(v1, registry)
	}

	return g
}

// initAIRegistry initializes the AI provider registry from configuration.
func initAIRegistry() *ai.Registry {
	credentials := facade.Config.AI.Credentials
	if len(credentials) == 0 {
		return nil
	}

	registry := ai.NewRegistry()

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
			provider, err = gemini.New(cfg)

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
