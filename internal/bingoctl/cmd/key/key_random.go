package key

import (
	"fmt"

	"github.com/bingo-project/component-base/cli/console"
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/bingo-project/component-base/util"
	"github.com/spf13/cobra"

	"github.com/bingo-project/bingo/internal/pkg/facade"
)

const (
	randomUsageStr = "random"
)

// RandomOptions is an option struct to support 'random' sub command.
type RandomOptions struct {
	// Options
	Length uint
}

// NewRandomOptions returns an initialized RandomOptions instance.
func NewRandomOptions() *RandomOptions {
	return &RandomOptions{}
}

// NewCmdRandom returns new initialized instance of 'random' sub command.
func NewCmdRandom() *cobra.Command {
	o := NewRandomOptions()

	cmd := &cobra.Command{
		Use:                   randomUsageStr,
		DisableFlagsInUseLine: true,
		Short:                 "A brief description of your command",
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
func (o *RandomOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

// Complete completes all the required options.
func (o *RandomOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// Run executes a new sub command using the specified options.
func (o *RandomOptions) Run(args []string) error {
	// Generate
	data := util.RandomString(int(o.Length))
	fmt.Println("key:", data)

	// Encrypt
	encrypt, err := facade.AES.EncryptString(data)
	console.ExitIf(err)

	fmt.Println("encrypt:", encrypt)

	return nil
}
