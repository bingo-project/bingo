package main

import (
	"github.com/spf13/cobra"

	"bingo/internal/bot"
)

func main() {
	command := bot.NewBotCommand()
	cobra.CheckErr(command.Execute())
}
