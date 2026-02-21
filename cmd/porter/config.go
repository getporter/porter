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
	cmd.AddCommand(buildConfigMigrateCommand(p))
	cmd.AddCommand(buildConfigContextCommands(p))

	return cmd
}

func buildConfigContextCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Context commands",
		Long:  "Commands for managing porter configuration contexts.",
	}

	cmd.AddCommand(buildConfigContextListCommand(p))
	cmd.AddCommand(buildConfigContextUseCommand(p))

	return cmd
}

func buildConfigContextListCommand(p *porter.Porter) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List configuration contexts",
		Long:    "List all contexts defined in the porter configuration file. The active context is marked with *.",
		Example: "  porter config context list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ConfigContextList(cmd.Context())
		},
	}
}

func buildConfigContextUseCommand(p *porter.Porter) *cobra.Command {
	return &cobra.Command{
		Use:     "use <name>",
		Short:   "Set the current configuration context",
		Long:    "Set the current-context in the porter configuration file.",
		Example: "  porter config context use prod",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ConfigContextUse(cmd.Context(), args[0])
		},
	}
}

func buildConfigMigrateCommand(p *porter.Porter) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the config file to the multi-context format",
		Long: `Migrate the porter config file from the legacy flat format to the
multi-context format (schemaVersion: "2.0.0"). The existing settings are
preserved under a context named "default".

Only YAML config files are supported for automatic migration. For TOML,
JSON, or HCL files, the required structure is printed so you can apply
the changes manually.`,
		Example: "  porter config migrate",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ConfigMigrate(cmd.Context())
		},
	}
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
