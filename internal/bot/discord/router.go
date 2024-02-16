package discord

import (
	"github.com/bwmarrin/discordgo"

	"bingo/internal/apiserver/store"
	"bingo/internal/bot/discord/controller/v1/server"
	"bingo/internal/bot/discord/middleware"
)

func RegisterCommandHandlers(s *discordgo.Session, i *discordgo.InteractionCreate) {
	middleware.Context(s, i)

	serverController := server.New(store.S, s, i)

	switch i.ApplicationCommandData().Name {

	// Ping
	case "ping":
		serverController.Pong()

	// Healthz
	case "healthz":
		serverController.Healthz()

	// Version
	case "version":
		serverController.Version()

	// Subscribe
	case "subscribe":
		serverController.Subscribe()

	// UnSubscribe
	case "unsubscribe":
		serverController.UnSubscribe()

	// Maintenance
	case "maintenance":
		serverController.ToggleMaintenance()

	default:
	}

	return
}
