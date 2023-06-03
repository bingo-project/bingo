package main

import (
	"github.com/spf13/cobra"

	"bingo/internal/apiserver"
)

func main() {
	command := apiserver.NewAppCommand()
	cobra.CheckErr(command.Execute())
}
