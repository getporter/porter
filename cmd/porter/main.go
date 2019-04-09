package main

import (
	"os"

	"github.com/gobuffalo/packr/v2"

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
		Use:   "porter",
		Short: "I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool",
		Example: `  porter create
  porter build
  porter install
  porter uninstall`,
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
	cmd.AddCommand(buildRunCommand(p))
	cmd.AddCommand(buildBundlesCommand(p))
	cmd.AddCommand(buildMixinsCommand(p))

	for _, alias := range buildBundleAliasCommands(p) {
		cmd.AddCommand(alias)
	}

	help := newHelptextBox()
	usage, _ := help.FindString("usage.txt")
	cmd.SetUsageTemplate(usage)

	return cmd
}

func newHelptextBox() *packr.Box {
	return packr.New("github.com/deislabs/porter/cmd/porter/helptext", "./helptext")
}
