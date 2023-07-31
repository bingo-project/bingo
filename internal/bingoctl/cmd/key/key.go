package key

import (
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/spf13/cobra"
)

const (
	aesUsageStr = "key"
)

// NewCmdKey returns new initialized instance of 'aes' sub command.
func NewCmdKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   aesUsageStr,
		DisableFlagsInUseLine: true,
		Short:                 "Key management",
		TraverseChildren:      true,
		Run:                   cmdutil.DefaultSubCommandRun(),
	}

	cmd.AddCommand(NewCmdRandom())
	cmd.AddCommand(NewCmdGenerate())
	cmd.AddCommand(NewCmdEncrypt())
	cmd.AddCommand(NewCmdDecrypt())

	return cmd
}
