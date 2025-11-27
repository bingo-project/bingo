package telegram

import (
	"gopkg.in/telebot.v3"

	"bingo/internal/pkg/store"
	"bingo/internal/bot/telegram/controller/v1/server"
	"bingo/internal/bot/telegram/middleware"
)

func RegisterRouters(b *telebot.Bot) {
	serverController := server.New(store.S)

	// Server
	b.Handle("/ping", serverController.Pong)
	b.Handle("/healthz", serverController.Healthz)
	b.Handle("/version", serverController.Version)
	b.Handle("/subscribe", serverController.Subscribe)
	b.Handle("/unsubscribe", serverController.UnSubscribe)

	// Admin
	adminOnly := b.Group()
	adminOnly.Use(middleware.AdminOnly)
	adminOnly.Handle("/maintenance", serverController.ToggleMaintenance)
}
