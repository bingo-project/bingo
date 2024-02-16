package discord

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bingo-project/component-base/log"
	"github.com/bwmarrin/discordgo"

	"bingo/internal/apiserver/facade"
)

func Run() {
	dg, err := discordgo.New("Bot " + facade.Config.Bot.Discord)
	if err != nil {
		log.Errorw("Error creating Discord session: " + err.Error())

		return
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		log.Errorw("Error opening Discord session: " + err.Error())

		return
	}

	log.Infow("Discord Bot started")

	// Register commands
	RegisterCommands(dg)

	// Register command handlers
	dg.AddHandler(RegisterCommandHandlers)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	_ = dg.Close()

	log.Infow("Discord Bot stopped")
}
