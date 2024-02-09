package apiserver

import (
	"time"

	"github.com/bingo-project/component-base/log"
	"gopkg.in/telebot.v3"

	"bingo/internal/apiserver/facade"
	"bingo/internal/apiserver/router"
)

// RunBot run bot server.
func RunBot() {
	pref := telebot.Settings{
		Token:  facade.Config.Bot.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalw("Failed to start bot: " + err.Error())
	}

	router.RegisterBotRouters(b)

	log.Infow("Bot server started")

	b.Start()
}
