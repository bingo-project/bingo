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

	// Healthz
	case "healthz":
		serverController.Healthz(s, m)

	// Version
	case "version":
		serverController.Version(s, m)

	// Maintenance
	case "maintenance":
		serverController.ToggleMaintenance(s, m)

	// Subscribe
	case "subscribe":
		serverController.Subscribe(s, m)

	// UnSubscribe
	case "unsubscribe":
		serverController.UnSubscribe(s, m)

	default:
	}

	return
}
