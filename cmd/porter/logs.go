package main

import (
	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildInstallationLogCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "logs",
		Aliases: []string{"log"},
		Short:   "Installation Logs commands",
		Long:    "Commands for working with installation logs",
	}
	cli.SetCommandGroup(cmd, "resource")

	cmd.AddCommand(buildInstallationLogShowCommand(p))

	return cmd
}

func buildInstallationLogShowCommand(p *porter.Porter) *cobra.Command {
	opts := &porter.LogsShowOptions{}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the logs from an installation",
		Long: `Show the logs from an installation.

Either display the logs from a specific run of a bundle with --run, or use --installation to display the logs from its most recent run.`,
		Example: `  porter installation logs show --installation wordpress --namespace dev
  porter installations logs show --run 01EZSWJXFATDE24XDHS5D5PWK6`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowInstallationLogs(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the installation is defined. Defaults to the global namespace.")
	f.StringVarP(&opts.Name, "installation", "i", "",
		"The installation that generated the logs.")
	f.StringVarP(&opts.RunID, "run", "r", "",
		"The bundle run that generated the logs.")

	return cmd
}
