package discord

import (
	"github.com/bwmarrin/discordgo"

	"bingo/internal/bot/discord/handler"
	"bingo/internal/bot/discord/middleware"
	"bingo/internal/pkg/store"
)

func RegisterCommandHandlers(s *discordgo.Session, i *discordgo.InteractionCreate) {
	middleware.Context(s, i)
	defer middleware.Recover()

	serverHandler := handler.NewServerHandler(store.S, s, i)

	switch i.ApplicationCommandData().Name {
	// Ping
	case "ping":
		serverHandler.Pong()

	// Healthz
	case "healthz":
		serverHandler.Healthz()

	// Version
	case "version":
		serverHandler.Version()

	// Subscribe
	case "subscribe":
		serverHandler.Subscribe()

	// UnSubscribe
	case "unsubscribe":
		serverHandler.UnSubscribe()

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
		serverHandler.ToggleMaintenance()

	default:
	}
}
