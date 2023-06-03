package user

import (
	"github.com/bingo-project/component-base/cli/genericclioptions"
	"github.com/bingo-project/component-base/cli/templates"
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/olekukonko/tablewriter"
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

// setHeader set headers for user commands.
func setHeader(table *tablewriter.Table) *tablewriter.Table {
	table.SetHeader([]string{"Username", "Nickname", "Email", "Phone", "Created At", "Updated At"})
	table.SetHeaderColor(tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgRedColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgMagentaColor},
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgWhiteColor})

	return table
}
