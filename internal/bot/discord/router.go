package discord

import (
	"github.com/bwmarrin/discordgo"

	"bingo/internal/admserver/store"
	"bingo/internal/bot/discord/controller/v1/server"
	"bingo/internal/bot/discord/middleware"
)

func RegisterCommandHandlers(s *discordgo.Session, i *discordgo.InteractionCreate) {
	middleware.Context(s, i)
	defer middleware.Recover()

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

	default:
	}

	// Admin only
	admin := middleware.IsAdmin(s, i)
	if !admin {
		return
	}

	switch i.ApplicationCommandData().Name {
	// Maintenance
	case "maintenance":
		serverController.ToggleMaintenance()

	default:
	}
}
