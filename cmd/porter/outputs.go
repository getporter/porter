package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildOutputsCommand(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "outputs",
		Aliases: []string{"output"},
		Short:   "Output commands",
		Long:    "Commands for working with Bundle outputs. TODO",
	}
	cmd.Annotations = map[string]string{
		"group": "resource",
	}

	cmd.AddCommand(buildOutputListCommand(p))
	cmd.AddCommand(buildOutputShowCommand(p))

	return cmd
}

func buildOutputListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.OutputListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List outputs of a bundle",
		Long:    "List bundle outputs. TODO",
		Example: `  porter outputs list -b BUNDLE [-o table|json|yaml]`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ListBundleOutputs(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")
	f.StringVarP(&opts.Bundle, "bundle", "b", "", "The bundle name to list outputs for.")

	return cmd
}

func buildOutputShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.OutputShowOptions{}

	cmd := &cobra.Command{
		Use:     "show",
		Short:   "Show a particular bundle output",
		Long:    "Show a bundle output. TODO",
		Example: `  porter output show -n OUTPUT -b BUNDLE [-o plaintext|json|yaml]`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowBundleOutput(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "plaintext",
		"Specify an output format.  Allowed values: plaintext, json, yaml")
	f.StringVarP(&opts.Output, "name", "n", "", "The output name.")
	f.StringVarP(&opts.Bundle, "bundle", "b", "", "The bundle name this output belongs to.")

	return cmd
}
