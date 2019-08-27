package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildVersionCommand(p *porter.Porter) *cobra.Command {
	opts := porter.VersionOptions{}
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the application version",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintVersion(opts)
		},
	}
	cmd.Annotations = map[string]string{
		"group": "meta",
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", string(porter.DefaultVersionFormat),
		"Specify an output format.  Allowed values: json, plaintext")

	return cmd
}
