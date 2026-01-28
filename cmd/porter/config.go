package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildConfigCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "config",
		Annotations: map[string]string{"group": "meta"},
		Short:       "Config commands",
		Long:        "Commands for managing Porter's configuration file.",
	}

	cmd.AddCommand(buildConfigShowCommand(p))
	cmd.AddCommand(buildConfigEditCommand(p))

	return cmd
}

func buildConfigShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ConfigShowOptions{}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the config file",
		Long:  "Show the contents of the porter configuration file.",
		Example: `  porter config show`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ConfigShow(cmd.Context(), opts)
		},
	}

	return cmd
}

func buildConfigEditCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ConfigEditOptions{}

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit the config file",
		Long: `Edit the porter configuration file.
If the config file does not exist, a default template will be created.

Uses the EDITOR environment variable to determine which editor to use.`,
		Example: `  porter config edit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ConfigEdit(cmd.Context(), opts)
		},
	}

	return cmd
}
