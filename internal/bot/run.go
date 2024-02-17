package bot

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bingo-project/component-base/log"

	"bingo/internal/apiserver/bootstrap"
	"bingo/internal/bot/discord"
	"bingo/internal/bot/telegram"
)

func run() error {
	bootstrap.Boot()

	telegramServer := telegram.NewTelegram()
	telegramServer.Run()

	discordServer := discord.NewDiscord()
	discordServer.Run()

	// 等待中断信号优雅地关闭服务器（10 秒超时)。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-quit
	log.Infow("Shutting down server ...")

	telegramServer.Close()
	discordServer.Close()

	return nil
}
