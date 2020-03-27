package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildInstanceCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "instances",
		Aliases: []string{"inst", "instance"},
		Short:   "Bundle Instance commands",
		Long:    "Commands for working with instances of a bundle",
	}
	cmd.Annotations = map[string]string{
		"group": "resource",
	}

	cmd.AddCommand(buildInstancesListCommand(p))
	cmd.AddCommand(buildInstanceShowCommand(p))

	cmd.AddCommand(buildInstanceOutputsCommands(p))

	return cmd
}

func buildInstancesListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List instances of installed bundles",
		Long: `List instances of all bundles installed by Porter.

A listing of instances of bundles currently installed by Porter will be provided, along with metadata such as creation time, last action, last status, etc.

Optional output formats include json and yaml.`,
		Example: `  porter instances list
  porter instances list -o json`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.ParseFormat()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ListInstances(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return cmd
}

func buildInstanceShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ShowOptions{}

	cmd := cobra.Command{
		Use:   "show [INSTANCE]",
		Short: "Show an instance of a bundle",
		Long:  "Displays info relating to an instance of a bundle, including status and a listing of outputs.",
		Example: `  porter instance show
porter instance show another-bundle

Optional output formats include json and yaml.
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowInstances(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return &cmd
}
