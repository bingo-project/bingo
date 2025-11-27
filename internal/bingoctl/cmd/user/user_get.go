package user

import (
	"context"
	"fmt"

	"github.com/bingo-project/component-base/cli/genericclioptions"
	"github.com/bingo-project/component-base/cli/templates"
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"bingo/internal/pkg/store"
	"bingo/pkg/store/where"
)

const (
	getUsageStr = "get USERNAME"
)

// GetOptions is an option struct to support 'get' sub command.
type GetOptions struct {
	Username string

	genericclioptions.IOStreams
}

var (
	getLong = templates.LongDesc(`
		# Get user foo detail information
		user get foo`)

	getExample = templates.Examples(`
		# Print all option values for get
		get arg1 arg2`)

	getUsageErrStr = fmt.Sprintf("expected '%s'.\nUSERNAME is required arguments for the get command", getUsageStr)
)

// NewGetOptions returns an initialized GetOptions instance.
func NewGetOptions(ioStreams genericclioptions.IOStreams) *GetOptions {
	return &GetOptions{
		IOStreams: ioStreams,
	}
}

// NewCmdGet returns new initialized instance of 'get' sub command.
func NewCmdGet(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewGetOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   getUsageStr,
		DisableFlagsInUseLine: true,
		Aliases:               []string{},
		Short:                 "A brief description of your command",
		TraverseChildren:      true,
		Long:                  getLong,
		Example:               getExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(args))
		},
		SuggestFor: []string{},
	}

	return cmd
}

// Complete completes all the required options.
func (o *GetOptions) Complete(cmd *cobra.Command, args []string) error {
	o.Username = args[0]

	return nil
}

// Validate makes sure there is no discrepancy in command options.
func (o *GetOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmdutil.UsageErrorf(cmd, getUsageErrStr)
	}

	return nil
}

// Run executes a get sub command using the specified options.
func (o *GetOptions) Run(args []string) error {
	whr := where.F("username", o.Username)
	user, err := store.S.User().Get(context.Background(), whr)
	if err != nil {
		return err
	}

	data := [][]string{
		{
			user.Username,
			user.Nickname,
			user.Email,
			user.Phone,
			user.CreatedAt.Format("2006-01-02 15:04:05"),
			user.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	table := tablewriter.NewWriter(o.Out)
	table = setHeader(table)
	table = cmdutil.TableWriterDefaultConfig(table)
	table.AppendBulk(data)
	table.Render()

	return nil
}
