package main

import (
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
			return p.PrintPlugins(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.RawFormat, "output", "o", "plaintext",
		"Output format, allowed values are: plaintext, json, yaml")

	return cmd
}

func buildPluginSearchCommand(p *porter.Porter) *cobra.Command {
	opts := porter.SearchOptions{
		Type: "plugin",
	}

	cmd := &cobra.Command{
		Use:   "search [QUERY]",
		Short: "Search available plugins",
		Long: `Search available plugins. You can specify an optional plugin name query, where the results are filtered by plugins whose name contains the query term.

By default the community plugin index at https://cdn.porter.sh/plugins/index.json is searched. To search from a mirror, set the environment variable PORTER_MIRROR, or mirror in the Porter config file, with the value to replace https://cdn.porter.sh with.`,
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

	flags := cmd.Flags()
	flags.StringVarP(&opts.RawFormat, "output", "o", "plaintext",
		"Output format, allowed values are: plaintext, json, yaml")
	flags.StringVar(&opts.Mirror, "mirror", pkgmgmt.DefaultPackageMirror,
		"Mirror of official Porter assets")

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
			return p.ShowPlugin(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.RawFormat, "output", "o", "plaintext",
		"Output format, allowed values are: plaintext, json, yaml")

	return cmd
}

func BuildPluginInstallCommand(p *porter.Porter) *cobra.Command {
	opts := plugins.InstallOptions{}
	cmd := &cobra.Command{
		Use:   "install NAME",
		Short: "Install a plugin",
		Long: `Install a plugin.

By default plugins are downloaded from the official Porter plugin feed at https://cdn.porter.sh/plugins/atom.xml. To download from a mirror, set the environment variable PORTER_MIRROR, or mirror in the Porter config file, with the value to replace https://cdn.porter.sh with.`,
		Example: `  porter plugin install azure  
  porter plugin install azure --url https://cdn.porter.sh/plugins/azure
  porter plugin install azure --feed-url https://cdn.porter.sh/plugins/atom.xml
  porter plugin install azure --version v0.8.2-beta.1
  porter plugin install azure --version canary`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InstallPlugin(cmd.Context(), opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Version, "version", "v", "latest",
		"The plugin version. This can either be a version number, or a tagged release like 'latest' or 'canary'")
	flags.StringVar(&opts.URL, "url", "",
		"URL from where the plugin can be downloaded, for example https://github.com/org/proj/releases/downloads")
	flags.StringVar(&opts.FeedURL, "feed-url", "",
		"URL of an atom feed where the plugin can be downloaded. Defaults to the official Porter plugin feed.")
	flags.StringVar(&opts.Mirror, "mirror", pkgmgmt.DefaultPackageMirror,
		"Mirror of official Porter assets")

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
			return p.UninstallPlugin(cmd.Context(), opts)
		},
	}

	return cmd
}

func buildPluginRunCommand(p *porter.Porter) *cobra.Command {
	var opts porter.RunInternalPluginOpts
	cmd := &cobra.Command{
		Use:   "run PLUGIN_KEY",
		Short: "Serve internal plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ApplyArgs(args)
			return p.RunInternalPlugins(cmd.Context(), opts)
		},
		Hidden: true, // This should ALWAYS be hidden, it is not a user-facing command
	}

	return cmd
}
