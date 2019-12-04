package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildPluginsCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugins",
		Aliases: []string{"plugin"},
		Hidden:  true,
		Short:   "Plugin commands. Plugins let Porter work on different cloud providers and systems.",
		Annotations: map[string]string{
			"group": "resource",
		},
	}

	cmd.AddCommand(buildPluginsListCommand(p))
	cmd.AddCommand(buildPluginRunCommand(p))

	return cmd
}

func buildPluginsListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.PrintPluginsOptions{}

	cmd := &cobra.Command{
		Use:    "list",
		Short:  "List installed plugins",
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.ParseFormat()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintPlugins(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.RawFormat, "output", "o", "table",
		"Output format, allowed values are: table, json, yaml")

	return cmd
}

func buildPluginRunCommand(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run KEY",
		Short: "Serve internal plugins",
		Run: func(cmd *cobra.Command, args []string) {
			p.RunInternalPlugins(args)
		},
		Hidden: true, // This should ALWAYS be hidden, it is not a user-facing command
	}

	return cmd
}
