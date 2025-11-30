package db

import (
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/spf13/cobra"

	"bingo/internal/bingoctl/database/seeder"
)

const (
	seedUsageStr = "seed"
)

// SeedOptions is an option struct to support 'seed' sub command.
type SeedOptions struct {
	Seeder string
}

// NewSeedOptions returns an initialized SeedOptions instance.
func NewSeedOptions() *SeedOptions {
	return &SeedOptions{}
}

// NewCmdSeed returns new initialized instance of 'seed' sub command.
func NewCmdSeed() *cobra.Command {
	o := NewSeedOptions()

	cmd := &cobra.Command{
		Use:                   seedUsageStr,
		DisableFlagsInUseLine: true,
		Short:                 "Init data",
		TraverseChildren:      true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Run(args))
		},
	}

	cmd.Flags().StringVarP(&o.Seeder, "seeder", "s", "", "Run specific seeder by signature")

	return cmd
}

// Validate makes sure there is no discrepancy in command options.
func (o *SeedOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

// Complete completes all the required options.
func (o *SeedOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// Run executes a new sub command using the specified options.
func (o *SeedOptions) Run(args []string) error {
	return seeder.RunSeeders(o.Seeder)
}
