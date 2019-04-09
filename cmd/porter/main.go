package main

import (
	"os"

	"github.com/deislabs/porter/pkg/porter"

	"github.com/spf13/cobra"
)

func main() {
	cmd := buildRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRootCommand() *cobra.Command {
	p := porter.New()
	cmd := &cobra.Command{
		Use:  "porter",
		Long: "I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Enable swapping out stdout/stderr for testing
			p.Out = cmd.OutOrStdout()
			p.Err = cmd.OutOrStderr()
		},
		SilenceUsage: true,
	}

	cmd.PersistentFlags().BoolVar(&p.Debug, "debug", false, "Enable debug logging")

	cmd.AddCommand(buildVersionCommand(p))
	cmd.AddCommand(buildSchemaCommand(p))
	cmd.AddCommand(buildCreateCommand(p))
	cmd.AddCommand(buildRunCommand(p))
	cmd.AddCommand(buildBuildCommand(p))
	cmd.AddCommand(buildBundlesCommand(p))

	for _, alias := range buildBundleAliasCommands(p) {
		cmd.AddCommand(alias)
	}


	return cmd
}

func buildListCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list resources",
	}

	cmd.AddCommand(buildListMixinsCommand(p))
	return cmd
}
