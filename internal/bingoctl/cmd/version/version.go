package version

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bingo-project/component-base/cli/genericclioptions"
	cmdutil "github.com/bingo-project/component-base/cli/util"
	"github.com/bingo-project/component-base/version"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

type Version struct {
	ClientVersion *version.Info `json:"clientVersion,omitempty" yaml:"clientVersion,omitempty"`
}

// Options is a struct to support version command.
type Options struct {
	Short  bool
	Output string

	genericclioptions.IOStreams
}

// NewOptions returns initialized Options.
func NewOptions(ioStreams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: ioStreams,
	}
}

// NewCmdVersion returns a cobra command for fetching versions.
func NewCmdVersion(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the client and server version information",
		Long:  "Print the client and server version information for the current context",
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(args))
		},
	}

	cmd.Flags().BoolVar(&o.Short, "short", o.Short, "If true, print just the version number.")
	cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "One of 'yaml' or 'json'.")

	return cmd
}

func (o *Options) Complete(cmd *cobra.Command, args []string) (err error) {
	return
}

// Validate makes sure there is no discrepancy in command options.
func (o *Options) Validate(cmd *cobra.Command, args []string) (err error) {
	if o.Output != "" && o.Output != "yaml" && o.Output != "json" {
		return errors.New(`--output must be 'yaml' or 'json'`)
	}

	return
}

// Run executes a creat subcommand using the specified options.
func (o *Options) Run(args []string) (err error) {
	var (
		serverErr   error
		versionInfo Version
	)

	clientVersion := version.Get()
	versionInfo.ClientVersion = &clientVersion

	switch o.Output {
	case "":
		if o.Short {
			fmt.Fprintf(o.Out, "Client Version: %s\n", clientVersion.GitVersion)
		} else {
			fmt.Fprintf(o.Out, "Client Version: %s\n", fmt.Sprintf("%#v", clientVersion))
		}
	case "yaml":
		marshaled, err := yaml.Marshal(&versionInfo)
		if err != nil {
			return err
		}

		fmt.Fprintln(o.Out, string(marshaled))
	case "json":
		marshaled, err := json.MarshalIndent(&versionInfo, "", "  ")
		if err != nil {
			return err
		}

		fmt.Fprintln(o.Out, string(marshaled))
	default:
		// There is a bug in the program if we hit this case.
		// However, we follow a policy of never panicking.
		return fmt.Errorf("VersionOptions were not validated: --output=%q should have been rejected", o.Output)
	}

	return serverErr
}
