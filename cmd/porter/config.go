package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildConfigCommand(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "config",
		Annotations: map[string]string{"group": "resource"},
		Short:       "Config commands",
		Long:        "Commands for managing Porter configuration",
	}

	cmd.AddCommand(buildConfigShowCommand(p))

	return cmd
}

func buildConfigShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ConfigShowOptions{}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show Porter configuration",
		Long:  "Display the current Porter configuration. If no config file exists, shows default values.",
		Example: `  porter config show
  porter config show -o json
  porter config show -o yaml
  porter config show -o toml`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowConfig(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "",
		"Output format (json, yaml, toml)")

	return cmd
}
