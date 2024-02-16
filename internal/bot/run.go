package bot

import (
	"bingo/internal/apiserver/bootstrap"
	"bingo/internal/bot/discord"
	"bingo/internal/bot/telegram"
)

func run() error {
	bootstrap.Boot()

	go telegram.Run()
	go discord.Run()

	select {}
}
