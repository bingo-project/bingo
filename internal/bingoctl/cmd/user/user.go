package user

import (
	"github.com/bingo-project/component-base/cli/genericclioptions"
	"github.com/bingo-project/component-base/cli/templates"
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/spf13/cobra"
)

var (
	userLong = templates.LongDesc(`User management commands.`)
)

// NewCmdUser returns new initialized instance of 'user' sub command.
func NewCmdUser(ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "user SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 "true",
		Long:                  userLong,
		Run:                   cmdutil.DefaultSubCommandRun(),
	}

	// add subcommands
	cmd.AddCommand(NewCmdList(ioStreams))
	cmd.AddCommand(NewCmdGet(ioStreams))

	return cmd
}
