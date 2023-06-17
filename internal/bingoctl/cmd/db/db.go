package db

import (
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/spf13/cobra"
)

const (
	dbUsageStr = "db"
)

// NewCmdDb returns new initialized instance of 'db' sub command.
func NewCmdDb() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   dbUsageStr,
		DisableFlagsInUseLine: true,
		Short:                 "Database management",
		TraverseChildren:      true,
		Run:                   cmdutil.DefaultSubCommandRun(),
	}

	cmd.AddCommand(NewCmdSeed())

	return cmd
}
