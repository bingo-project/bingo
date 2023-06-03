package main

import (
	"github.com/spf13/cobra"

	"bingo/internal/bingoctl/cmd"
)

func main() {
	command := cmd.NewDefaultBingoCtlCommand()
	cobra.CheckErr(command.Execute())
}
