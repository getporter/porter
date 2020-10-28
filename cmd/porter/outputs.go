package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildInstallationOutputsCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "output",
		Aliases: []string{"outputs"},
		Short:   "Output commands",
		Annotations: map[string]string{
			"group": "resource",
		},
	}

	cmd.AddCommand(buildBundleOutputShowCommand(p))
	cmd.AddCommand(buildBundleOutputListCommand(p))

	return cmd
}

func buildBundleOutputListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.OutputListOptions{}

	cmd := cobra.Command{
		Use:   "list [--installation|i INSTALLATION]",
		Short: "List installation outputs",
		Long:  "Displays a listing of installation outputs.",
		Example: `  porter installation outputs list
    porter installation outputs list --installation another-bundle
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintBundleOutputs(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")
	f.StringVarP(&opts.Name, "installation", "i", "",
		"Specify the installation to which the output belongs.")

	return &cmd
}

func buildBundleOutputShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.OutputShowOptions{}

	cmd := cobra.Command{
		Use:   "show NAME [--installation|-i INSTALLATION]",
		Short: "Show the output of an installation",
		Long:  "Show the output of an installation",
		Example: `  porter installation output show kubeconfig
    porter installation output show subscription-id --installation azure-mysql`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowBundleOutput(&opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Name, "installation", "i", "",
		"Specify the installation to which the output belongs.")

	return &cmd
}
