package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildInstallationCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "installations",
		Aliases: []string{"inst", "installation"},
		Short:   "Installation commands",
		Long:    "Commands for working with installations of a bundle",
	}
	cmd.Annotations = map[string]string{
		"group": "resource",
	}

	cmd.AddCommand(buildInstallationsListCommand(p))
	cmd.AddCommand(buildInstallationShowCommand(p))
	cmd.AddCommand(buildInstallationOutputsCommands(p))
	cmd.AddCommand(buildInstallationDeleteCommand(p))

	return cmd
}

func buildInstallationsListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed bundles",
		Long: `List all bundles installed by Porter.

A listing of bundles currently installed by Porter will be provided, along with metadata such as creation time, last action, last status, etc.

Optional output formats include json and yaml.`,
		Example: `  porter installations list
  porter installations list -o json`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.ParseFormat()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintInstallations(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return cmd
}

func buildInstallationShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ShowOptions{}

	cmd := cobra.Command{
		Use:   "show [INSTALLATION]",
		Short: "Show an installation of a bundle",
		Long:  "Displays info relating to an installation of a bundle, including status and a listing of outputs.",
		Example: `  porter installation show
porter installation show another-bundle

Optional output formats include json and yaml.
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowInstallation(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return &cmd
}

func buildInstallationDeleteCommand(p *porter.Porter) *cobra.Command {
	opts := porter.DeleteOptions{}

	cmd := cobra.Command{
		Use:   "delete [INSTALLATION]",
		Short: "Delete an installation",
		Long:  "Deletes an installation, including all claim, result and output records.",
		Example: `  porter installation delete
porter installation delete another-installation
porter installation delete --force
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.DeleteInstallation(opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.Force, "force", false,
		"Force a delete the installation, regardless of last completed action")

	return &cmd
}
