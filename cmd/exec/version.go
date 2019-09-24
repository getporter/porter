package main

import (
	"github.com/deislabs/porter/pkg/exec"
	"github.com/deislabs/porter/pkg/porter/version"
	"github.com/spf13/cobra"
)

func buildVersionCommand(m *exec.Mixin) *cobra.Command {
	opts := version.VersionOpts{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the mixin version",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.PrintVersion(opts.Options)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", string(version.DefaultVersionFormat),
		"Specify an output format.  Allowed values: json, plaintext")

	return cmd
}
