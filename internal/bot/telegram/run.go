package telegram

import (
	"os"
	"os/signal"
	"syscall"
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

	go b.Start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	b.Stop()

	log.Infow("Telegram Bot stopped")
}
