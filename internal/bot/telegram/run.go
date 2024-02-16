package telegram

import (
	"time"

	"github.com/bingo-project/component-base/log"
	"gopkg.in/telebot.v3"

	"bingo/internal/apiserver/facade"
	"bingo/internal/bot/telegram/middleware"
)

func Run() {
	pref := telebot.Settings{
		Token:  facade.Config.Bot.Telegram,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalw("Failed to start bot: " + err.Error())
	}

	// Global middleware
	b.Use(middleware.Context)
	b.Use(middleware.Recover)

	RegisterRouters(b)

	log.Infow("Telegram Bot started")

	b.Start()
}
