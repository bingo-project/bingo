package telegram

import (
	"time"

	"github.com/bingo-project/component-base/log"
	"gopkg.in/telebot.v3"

	"bingo/internal/apiserver/facade"
	"bingo/internal/bot/telegram/middleware"
)

type TelegramServer struct {
	*telebot.Bot
}

func NewTelegram() *TelegramServer {
	pref := telebot.Settings{
		Token:  facade.Config.Bot.Telegram,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalw("Failed to start bot: " + err.Error())
	}

	return &TelegramServer{b}
}

func (b *TelegramServer) Run() {

	// Global middleware
	b.Use(middleware.Context)
	b.Use(middleware.Recover)

	RegisterRouters(b.Bot)

	log.Infow("Telegram Bot started")

	go b.Start()
}

func (b *TelegramServer) Close() {
	b.Stop()

	log.Infow("Telegram Bot stopped")
}
