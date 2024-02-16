package discord

import (
	"github.com/bwmarrin/discordgo"

	"bingo/internal/apiserver/store"
	"bingo/internal/bot/discord/controller/v1/server"
	"bingo/internal/bot/discord/middleware"
)

func RegisterCommandHandlers(s *discordgo.Session, i *discordgo.InteractionCreate) {
	middleware.Context(s, i)

	serverController := server.New(store.S)

	switch i.ApplicationCommandData().Name {

	// Ping
	case "ping":
		serverController.Pong(s, i)

	// Healthz
	case "healthz":
		serverController.Healthz(s, i)

	// Version
	case "version":
		serverController.Version(s, i)

	// Subscribe
	case "subscribe":
		serverController.Subscribe(s, i)

	// UnSubscribe
	case "unsubscribe":
		serverController.UnSubscribe(s, i)

	// Maintenance
	case "maintenance":
		serverController.ToggleMaintenance(s, i)

	default:
	}

	return
}
