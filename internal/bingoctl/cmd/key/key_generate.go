package key

import (
	"fmt"

	"github.com/bingo-project/component-base/cli/console"
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/bingo-project/component-base/util"
	"github.com/spf13/cobra"

	"bingo/internal/apiserver/facade"
)

const (
	generateUsageStr = "generate"
)

// GenerateOptions is an option struct to support 'generate' sub command.
type GenerateOptions struct {
	// Options
	Length uint
}

// NewGenerateOptions returns an initialized GenerateOptions instance.
func NewGenerateOptions() *GenerateOptions {
	return &GenerateOptions{}
}

// NewCmdGenerate returns new initialized instance of 'generate' sub command.
func NewCmdGenerate() *cobra.Command {
	o := NewGenerateOptions()

	cmd := &cobra.Command{
		Use:                   generateUsageStr,
		DisableFlagsInUseLine: true,
		Short:                 "Generate a key",
		TraverseChildren:      true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Run(args))
		},
	}

	cmd.Flags().UintVarP(&o.Length, "length", "l", 32, "Key length.")

	return cmd
}

// Validate makes sure there is no discrepancy in command options.
func (o *GenerateOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

// Complete completes all the required options.
func (o *GenerateOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// Run executes a new sub command using the specified options.
func (o *GenerateOptions) Run(args []string) error {
	// Generate
	data := util.RandomString(int(o.Length))
	fmt.Println("key:", data)

	// Encrypt
	encrypt, err := facade.AES.EncryptString(data)
	console.ExitIf(err)

	fmt.Println("encrypt:", encrypt)

	return nil
}
