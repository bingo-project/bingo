// ABOUTME: Chat router registration for AI endpoints.
// ABOUTME: Registers chat completions, models, and session routes.

package router

import (
	"github.com/gin-gonic/gin"

	chathandler "github.com/bingo-project/bingo/internal/apiserver/handler/http/chat"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ai"
)

// MapChatRouters registers AI chat routes
func MapChatRouters(g *gin.RouterGroup, registry *ai.Registry) {
	chatHandler := chathandler.NewChatHandler(store.S, registry)
	sessionHandler := chathandler.NewSessionHandler(store.S, registry)

	// OpenAI-compatible endpoints
	g.POST("/chat/completions", chatHandler.ChatCompletions)
	g.GET("/models", chatHandler.ListModels)

	// Session management
	sessions := g.Group("/ai/sessions")
	{
		sessions.POST("", sessionHandler.CreateSession)
		sessions.GET("", sessionHandler.ListSessions)
		sessions.GET("/:session_id", sessionHandler.GetSession)
		sessions.DELETE("/:session_id", sessionHandler.DeleteSession)
		sessions.GET("/:session_id/history", sessionHandler.GetSessionHistory)
	}
}
