package cmd

import (
	"io"
	"os"

	"github.com/bingo-project/component-base/cli/genericclioptions"
	"github.com/bingo-project/component-base/cli/templates"
	"github.com/bingo-project/component-base/cmd/options"
	"github.com/spf13/cobra"

	"github.com/bingo-project/bingo/internal/bingoctl/cmd/key"
	"github.com/bingo-project/bingo/internal/bingoctl/cmd/user"
	"github.com/bingo-project/bingo/internal/bingoctl/cmd/version"
	"github.com/bingo-project/bingo/internal/pkg/bootstrap"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

func NewDefaultBingoCtlCommand() *cobra.Command {
	return NewBingoCtlCommand(os.Stdin, os.Stdout, os.Stderr)
}

func NewBingoCtlCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	var cmds = &cobra.Command{
		Use:   "bingoctl",
		Short: "bingoctl is the bingo startup client",
		Long:  `bingoctl is the client side tool for bingo startup.`,
		Run:   runHelp,
	}

	// Load config
	// cobra.OnInitialize(initConfig)
	initConfig()

	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: err}

	groups := templates.CommandGroups{
		{
			Message: "Tool Commands:",
			Commands: []*cobra.Command{
				key.NewCmdKey(),
			},
		},
		{
			Message: "Advanced Commands:",
			Commands: []*cobra.Command{
				user.NewCmdUser(ioStreams),
			},
		},
	}
	groups.Add(cmds)

	filters := []string{""}
	templates.ActsAsRootCommand(cmds, filters, groups...)

	// Config file
	cmds.PersistentFlags().StringVarP(&bootstrap.CfgFile, "config", "c", "", "The path to the configuration file. Empty string for no configuration file.")

	// Add commands
	cmds.AddCommand(version.NewCmdVersion(ioStreams))
	cmds.AddCommand(options.NewCmdOptions())

	return cmds
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	bootstrap.InitConfig("bingoctl.yaml")
	bootstrap.Boot()

	// Init store
	_ = store.NewStore(bootstrap.InitDB())
}

func runHelp(cmd *cobra.Command, args []string) {
	_ = cmd.Help()
}
