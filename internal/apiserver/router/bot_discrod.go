package router

import (
	"github.com/bwmarrin/discordgo"

	"bingo/internal/apiserver/bot/discord/controller/v1/server"
	"bingo/internal/apiserver/store"
)

func RegisterBotDiscordRouters(s *discordgo.Session, m *discordgo.MessageCreate) {
	serverController := server.New(store.S)

	switch m.Content {

	// Ping
	case "ping":
		serverController.Pong(s, m)

	// Status
	case "healthz":
		serverController.Healthz(s, m)

	default:
	}

	return
}
