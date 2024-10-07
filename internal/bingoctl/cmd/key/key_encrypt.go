package key

import (
	"fmt"

	"github.com/bingo-project/component-base/cli/console"
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/spf13/cobra"

	"bingo/internal/pkg/facade"
)

const (
	encryptUsageStr = "encrypt STRING"
)

var (
	encryptUsageErrStr = fmt.Sprintf(
		"expected '%s'.\nNAME is a required argument for the encrypt command",
		encryptUsageStr,
	)
)

// EncryptOptions is an option struct to support 'encrypt' sub command.
type EncryptOptions struct {
	// Options
	Str string
}

// NewEncryptOptions returns an initialized EncryptOptions instance.
func NewEncryptOptions() *EncryptOptions {
	return &EncryptOptions{}
}

// NewCmdEncrypt returns new initialized instance of 'encrypt' sub command.
func NewCmdEncrypt() *cobra.Command {
	o := NewEncryptOptions()

	cmd := &cobra.Command{
		Use:                   encryptUsageStr,
		DisableFlagsInUseLine: true,
		Short:                 "Encrypt string",
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
func (o *EncryptOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmdutil.UsageErrorf(cmd, encryptUsageErrStr)
	}

	return nil
}

// Complete completes all the required options.
func (o *EncryptOptions) Complete(cmd *cobra.Command, args []string) error {
	o.Str = args[0]

	return nil
}

// Run executes a new sub command using the specified options.
func (o *EncryptOptions) Run(args []string) error {
	data, err := facade.AES.EncryptString(o.Str)
	console.ExitIf(err)

	fmt.Println("encrypt:", data)

	return nil
}
