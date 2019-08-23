package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildInstanceOutputsCommands(p *porter.Porter) *cobra.Command {
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
		Use:   "list [--instance|i INSTANCE]",
		Short: "List bundle instance outputs",
		Long:  "Displays a listing of bundle instance outputs.",
		Example: `  porter instance outputs list
    porter instance outputs list --instance another-bundle
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ListBundleOutputs(&opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")
	f.StringVarP(&opts.Name, "instance", "i", "",
		"Specify the bundle instance to which the output belongs.")

	return &cmd
}

func buildBundleOutputShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.OutputShowOptions{}

	cmd := cobra.Command{
		Use:   "show NAME [--instance|-i INSTANCE]",
		Short: "Show the output of a bundle instance",
		Long:  "Show the output of a bundle instance",
		Example: `  porter instance output show kubeconfig
    porter instance output show subscription-id --instance azure-mysql`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowBundleOutput(&opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Name, "instance", "i", "",
		"Specify the bundle instance to which the output belongs.")

	return &cmd
}
