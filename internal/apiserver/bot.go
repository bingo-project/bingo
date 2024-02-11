package apiserver

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bingo-project/component-base/log"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/telebot.v3"

	"bingo/internal/apiserver/facade"
	"bingo/internal/apiserver/router"
)

// RunBotTelegram run bot server.
func RunBotTelegram() {
	pref := telebot.Settings{
		Token:  facade.Config.Bot.Telegram,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalw("Failed to start bot: " + err.Error())
	}

	router.RegisterBotRouters(b)

	log.Infow("Telegram Bot started")

	b.Start()
}

// RunBotDiscord run discord bot server.
func RunBotDiscord() {
	dg, err := discordgo.New("Bot " + facade.Config.Bot.Discord)
	if err != nil {
		log.Errorw("Error creating Discord session: " + err.Error())

		return
	}

	dg.AddHandler(router.RegisterBotDiscordRouters)

	// dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		log.Errorw("Error opening Discord session: " + err.Error())

		return
	}

	log.Infow("Discord Bot started")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	_ = dg.Close()

	log.Infow("Discord Bot stopped")
}
