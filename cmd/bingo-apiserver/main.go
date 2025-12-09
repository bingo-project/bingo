package main

import (
	"github.com/spf13/cobra"

	"github.com/bingo-project/bingo/internal/apiserver"
)

func main() {
	command := apiserver.NewAppCommand()
	cobra.CheckErr(command.Execute())
}
