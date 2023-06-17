package key

import (
	"fmt"

	"github.com/bingo-project/component-base/cli/console"
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/spf13/cobra"

	"bingo/internal/apiserver/facade"
)

const (
	decryptUsageStr = "decrypt STRING"
)

var (
	decryptUsageErrStr = fmt.Sprintf(
		"expected '%s'.\nNAME is a required argument for the decrypt command",
		decryptUsageStr,
	)
)

// DecryptOptions is an option struct to support 'decrypt' sub command.
type DecryptOptions struct {
	// Options
	Str string
}

// NewDecryptOptions returns an initialized DecryptOptions instance.
func NewDecryptOptions() *DecryptOptions {
	return &DecryptOptions{}
}

// NewCmdDecrypt returns new initialized instance of 'decrypt' sub command.
func NewCmdDecrypt() *cobra.Command {
	o := NewDecryptOptions()

	cmd := &cobra.Command{
		Use:                   decryptUsageStr,
		DisableFlagsInUseLine: true,
		Short:                 "Decrypt string",
		TraverseChildren:      true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Run(args))
		},
	}

	return cmd
}

// Validate makes sure there is no discrepancy in command options.
func (o *DecryptOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmdutil.UsageErrorf(cmd, decryptUsageErrStr)
	}

	return nil
}

// Complete completes all the required options.
func (o *DecryptOptions) Complete(cmd *cobra.Command, args []string) error {
	o.Str = args[0]

	return nil
}

// Run executes a new sub command using the specified options.
func (o *DecryptOptions) Run(args []string) error {
	data, err := facade.AES.DecryptString(o.Str)
	console.ExitIf(err)

	fmt.Println("decrypt:", data)

	return nil
}
