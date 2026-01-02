// ABOUTME: AI router registration for chat, session, and agent endpoints.
// ABOUTME: Registers chat completions, models, sessions, and agent preset routes.

package router

import (
	"github.com/gin-gonic/gin"

	bizauth "github.com/bingo-project/bingo/internal/apiserver/biz/auth"
	chathandler "github.com/bingo-project/bingo/internal/apiserver/handler/http/chat"
	"github.com/bingo-project/bingo/internal/apiserver/middleware"
	"github.com/bingo-project/bingo/internal/pkg/ai"
	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	httpmw "github.com/bingo-project/bingo/internal/pkg/middleware/http"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

// MapAiRouters registers AI-related routes (chat, sessions, roles)
func MapAiRouters(g *gin.Engine) {
	// Use global registry
	registry := ai.GetRegistry()
	if registry == nil {
		return
	}

	// v1 group
	v1 := g.Group("/v1")
	v1.Use(middleware.Lang())
	v1.Use(middleware.Maintenance())

	// Authentication middleware
	loader := bizauth.NewUserLoader(store.S)
	authn := auth.New(loader)
	v1.Use(auth.Middleware(authn))

	// Initialize handlers
	chatHandler := chathandler.NewChatHandler(store.S, registry)
	sessionHandler := chathandler.NewSessionHandler(store.S, registry)
	agentHandler := chathandler.NewAgentHandler(store.S)

	// Get AI quota limit
	rpm := facade.Config.AI.Quota.DefaultRPM
	if rpm <= 0 {
		rpm = 20 // fallback default
	}

	// OpenAI-compatible endpoints
	// Apply rate limiter only to chat completions (consumes quota)
	v1.POST("/chat/completions", httpmw.AILimiter(rpm), chatHandler.ChatCompletions)
	v1.GET("/models", chatHandler.ListModels)

	// Session management
	sessions := v1.Group("/ai/sessions")
	{
		sessions.POST("", sessionHandler.CreateSession)
		sessions.GET("", sessionHandler.ListSessions)
		sessions.GET("/:session_id", sessionHandler.GetSession)
		sessions.PUT("/:session_id", sessionHandler.UpdateSession)
		sessions.DELETE("/:session_id", sessionHandler.DeleteSession)
		sessions.GET("/:session_id/history", sessionHandler.GetSessionHistory)
	}

	// Agent presets (read-only for users)
	agents := v1.Group("/ai/agents")
	{
		agents.GET("", agentHandler.List)
		agents.GET("/:id", agentHandler.Get)
	}
}
