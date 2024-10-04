package main

import (
	"github.com/spf13/cobra"

	"bingo/internal/scheduler"
)

func main() {
	command := scheduler.NewSchedulerCommand()
	cobra.CheckErr(command.Execute())
}
