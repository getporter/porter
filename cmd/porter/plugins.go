package main

import (
	"fmt"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildPluginsCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugins",
		Aliases: []string{"plugin"},
		Short:   "Plugin commands. Plugins enable Porter to work on different cloud providers and systems.",
		Annotations: map[string]string{
			"group": "resource",
		},
	}

	cmd.AddCommand(buildPluginsListCommand(p))
	cmd.AddCommand(buildPluginSearchCommand(p))
	cmd.AddCommand(buildPluginShowCommand(p))
	cmd.AddCommand(BuildPluginInstallCommand(p))
	cmd.AddCommand(BuildPluginUninstallCommand(p))
	cmd.AddCommand(buildPluginRunCommand(p))

	return cmd
}

func buildPluginsListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.PrintPluginsOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
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

func buildPluginSearchCommand(p *porter.Porter) *cobra.Command {
	opts := porter.SearchOptions{
		Type: "plugin",
	}

	cmd := &cobra.Command{
		Use:   "search [QUERY]",
		Short: "Search available plugins",
		Long:  "Search available plugins. You can specify an optional plugin name query, where the results are filtered by plugins whose name contains the query term.",
		Example: `  porter plugin search
  porter plugin search azure
  porter plugin search -o json`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.SearchPackages(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.RawFormat, "output", "o", "table",
		"Output format, allowed values are: table, json, yaml")

	return cmd
}

func buildPluginShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ShowPluginOptions{}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show details about an installed plugin",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowPlugin(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.RawFormat, "output", "o", "table",
		"Output format, allowed values are: table, json, yaml")

	return cmd
}

func BuildPluginInstallCommand(p *porter.Porter) *cobra.Command {
	opts := plugins.InstallOptions{}
	cmd := &cobra.Command{
		Use:   "install NAME",
		Short: "Install a plugin",
		Example: `  porter plugin install azure  
  porter plugin install azure --url https://cdn.porter.sh/plugins/azure
  porter plugin install azure --feed-url https://cdn.porter.sh/plugins/atom.xml
  porter plugin install azure --version v0.8.2-beta.1
  porter plugin install azure --version canary`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InstallPlugin(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Version, "version", "v", "latest",
		"The plugin version. This can either be a version number, or a tagged release like 'latest' or 'canary'")
	cmd.Flags().StringVar(&opts.URL, "url", "",
		"URL from where the plugin can be downloaded, for example https://github.com/org/proj/releases/downloads")
	cmd.Flags().StringVar(&opts.FeedURL, "feed-url", "",
		fmt.Sprintf(`URL of an atom feed where the plugin can be downloaded (default %s)`, plugins.DefaultFeedUrl))
	return cmd
}

func BuildPluginUninstallCommand(p *porter.Porter) *cobra.Command {
	opts := pkgmgmt.UninstallOptions{}
	cmd := &cobra.Command{
		Use:     "uninstall NAME",
		Short:   "Uninstall a plugin",
		Example: `  porter plugin uninstall azure`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.UninstallPlugin(opts)
		},
	}

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
