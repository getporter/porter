package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/deislabs/porter/pkg/porter/version"
	"github.com/spf13/cobra"
)

func buildVersionCommand(p *porter.Porter) *cobra.Command {
	opts := version.VersionOpts{}
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the application version",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Options.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.System {
				return p.PrintDebugInfo(opts.Options)
			}
			return p.PrintVersion(opts.Options)
		},
	}
	cmd.Annotations = map[string]string{
		"group": "meta",
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", string(version.DefaultVersionFormat),
		"Specify an output format.  Allowed values: json, plaintext")
	f.BoolVarP(&opts.System, "system", "s", false, "Print system debug information")

	return cmd
}
