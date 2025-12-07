package telegram

import (
	"gopkg.in/telebot.v3"

	"bingo/internal/bot/telegram/handler"
	"bingo/internal/bot/telegram/middleware"
	"bingo/internal/pkg/store"
)

func RegisterRouters(b *telebot.Bot) {
	serverHandler := handler.NewServerHandler(store.S)

	// Server
	b.Handle("/ping", serverHandler.Pong)
	b.Handle("/healthz", serverHandler.Healthz)
	b.Handle("/version", serverHandler.Version)
	b.Handle("/subscribe", serverHandler.Subscribe)
	b.Handle("/unsubscribe", serverHandler.UnSubscribe)

	// Admin
	adminOnly := b.Group()
	adminOnly.Use(middleware.AdminOnly)
	adminOnly.Handle("/maintenance", serverHandler.ToggleMaintenance)
}
