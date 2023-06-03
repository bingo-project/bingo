package cmd

import (
	"io"
	"os"

	"github.com/bingo-project/bingoctl/pkg/cmd/create"
	"github.com/bingo-project/bingoctl/pkg/cmd/gen"
	makecmd "github.com/bingo-project/bingoctl/pkg/cmd/make"
	"github.com/bingo-project/bingoctl/pkg/cmd/version"
	"github.com/bingo-project/component-base/cli/genericclioptions"
	"github.com/bingo-project/component-base/cli/templates"
	"github.com/spf13/cobra"

	"bingo/internal/apiserver"
	"bingo/internal/bingoctl/cmd/migrate"
	"bingo/internal/bingoctl/cmd/user"
	"bingo/internal/pkg/log"
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
	cobra.OnInitialize(initConfig)

	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: err}

	groups := templates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				makecmd.NewCmdMake(),
				create.NewCmdCreate(),
				gen.NewCmdGen(),
				migrate.NewCmdMigrate(),
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
	cmds.PersistentFlags().StringVarP(&apiserver.CfgFile, "config", "c", "", "The path to the configuration file. Empty string for no configuration file.")

	// Add commands
	cmds.AddCommand(version.NewCmdVersion(ioStreams))

	return cmds
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	apiserver.InitConfig()

	// Init store
	if err := apiserver.InitStore(); err != nil {
		log.Fatalw(err.Error())
	}
}

func runHelp(cmd *cobra.Command, args []string) {
	_ = cmd.Help()
}
