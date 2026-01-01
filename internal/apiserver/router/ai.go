// ABOUTME: AI router registration for chat, session, and role endpoints.
// ABOUTME: Registers chat completions, models, sessions, and role preset routes.

package router

import (
	"github.com/gin-gonic/gin"

	chathandler "github.com/bingo-project/bingo/internal/apiserver/handler/http/chat"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	httpmw "github.com/bingo-project/bingo/internal/pkg/middleware/http"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ai"
)

// MapAiRouters registers AI-related routes (chat, sessions, roles)
func MapAiRouters(g *gin.RouterGroup, registry *ai.Registry) {
	chatHandler := chathandler.NewChatHandler(store.S, registry)
	sessionHandler := chathandler.NewSessionHandler(store.S, registry)
	roleHandler := chathandler.NewRoleHandler(store.S, registry)

	// Get AI quota limit
	rpm := facade.Config.AI.Quota.DefaultRPM
	if rpm <= 0 {
		rpm = 20 // fallback default
	}

	// OpenAI-compatible endpoints
	// Apply rate limiter only to chat completions (consumes quota)
	g.POST("/chat/completions", httpmw.AILimiter(rpm), chatHandler.ChatCompletions)
	g.GET("/models", chatHandler.ListModels)

	// Session management
	sessions := g.Group("/ai/sessions")
	{
		sessions.POST("", sessionHandler.CreateSession)
		sessions.GET("", sessionHandler.ListSessions)
		sessions.GET("/:session_id", sessionHandler.GetSession)
		sessions.PUT("/:session_id", sessionHandler.UpdateSession)
		sessions.DELETE("/:session_id", sessionHandler.DeleteSession)
		sessions.GET("/:session_id/history", sessionHandler.GetSessionHistory)
	}

	// Role presets (read-only for users)
	roles := g.Group("/ai/roles")
	{
		roles.GET("", roleHandler.List)
		roles.GET("/:role_id", roleHandler.Get)
	}
}
