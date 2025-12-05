package key

import (
	"os"
	"strings"

	"github.com/bingo-project/component-base/cli/console"
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/bingo-project/component-base/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"bingo/internal/pkg/facade"
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
	key := util.RandomString(int(o.Length))
	err := o.writeNewEnvironmentFileWith(key)
	if err != nil {
		console.ExitIf(err)
	}

	console.Info("key set successfully")

	return nil
}

// writeNewEnvironmentFileWith Write a new environment file with the given key.
func (o *GenerateOptions) writeNewEnvironmentFileWith(key string) error {
	path := viper.ConfigFileUsed()
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newContent := strings.Replace(string(content), "key: "+facade.Config.App.Key, "key: "+key, 1)

	err = os.WriteFile(path, []byte(newContent), 0600)
	if err != nil {
		return err
	}

	return nil
}
