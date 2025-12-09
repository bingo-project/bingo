package main

import (
	"github.com/spf13/cobra"

	"github.com/bingo-project/bingo/internal/admserver"
)

func main() {
	command := admserver.NewAppCommand()
	cobra.CheckErr(command.Execute())
}
