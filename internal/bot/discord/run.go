package discord

import (
	"github.com/bingo-project/component-base/log"
	"github.com/bwmarrin/discordgo"

	"bingo/internal/apiserver/facade"
)

type DiscordServer struct {
	*discordgo.Session
}

func NewDiscord() *DiscordServer {
	dg, err := discordgo.New("Bot " + facade.Config.Bot.Discord)
	if err != nil {
		log.Errorw("Error creating Discord session: " + err.Error())

		return nil
	}

	return &DiscordServer{dg}
}

func (s *DiscordServer) Run() {
	s.Identify.Intents = discordgo.IntentsGuildMessages

	err := s.Open()
	if err != nil {
		log.Errorw("Error opening Discord session: " + err.Error())

		return
	}

	log.Infow("Discord Bot started")

	// Register commands
	RegisterCommands(s.Session)

	// Register command handlers
	s.AddHandler(RegisterCommandHandlers)
}

func (s *DiscordServer) Close() {
	_ = s.Session.Close()

	log.Infow("Discord Bot stopped")
}
