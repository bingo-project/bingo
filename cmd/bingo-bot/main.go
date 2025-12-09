package main

import (
	"github.com/spf13/cobra"

	"github.com/bingo-project/bingo/internal/bot"
)

func main() {
	command := bot.NewBotCommand()
	cobra.CheckErr(command.Execute())
}
