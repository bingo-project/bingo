package router

import (
	"gopkg.in/telebot.v3"

	"bingo/internal/apiserver/bot/controller/v1/server"
	"bingo/internal/apiserver/store"
)

func RegisterBotRouters(b *telebot.Bot) {
	serverController := server.New(store.S)

	// Server
	b.Handle("/healthz", serverController.Healthz)
	b.Handle("/version", serverController.Version)
}
