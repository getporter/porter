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
	cmd.AddCommand(buildCredentialsCommand(p))
	cmd.AddCommand(buildOutputsCommand(p))

	for _, alias := range buildBundleAliasCommands(p) {
		cmd.AddCommand(alias)
	}

	// Hide the help command from the help text
	cmd.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Hidden: true,
	})

	help := newHelptextBox()
	usage, _ := help.FindString("usage.txt")
	cmd.SetUsageTemplate(usage)
	cobra.AddTemplateFunc("ShouldShowGroupCommands", ShouldShowGroupCommands)
	cobra.AddTemplateFunc("ShouldShowGroupCommand", ShouldShowGroupCommand)
	cobra.AddTemplateFunc("ShouldShowUngroupedCommands", ShouldShowUngroupedCommands)
	cobra.AddTemplateFunc("ShouldShowUngroupedCommand", ShouldShowUngroupedCommand)

	return cmd
}

func newHelptextBox() *packr.Box {
	return packr.New("github.com/deislabs/porter/cmd/porter/helptext", "./helptext")
}

func ShouldShowGroupCommands(cmd *cobra.Command, group string) bool {
	for _, child := range cmd.Commands() {
		if ShouldShowGroupCommand(child, group) {
			return true
		}
	}
	return false
}

func ShouldShowGroupCommand(cmd *cobra.Command, group string) bool {
	if cmd.Annotations["group"] == group {
		return true
	}
	return false
}

func ShouldShowUngroupedCommands(cmd *cobra.Command) bool {
	for _, child := range cmd.Commands() {
		if ShouldShowUngroupedCommand(child) {
			return true
		}
	}
	return false
}

func ShouldShowUngroupedCommand(cmd *cobra.Command) bool {
	if !cmd.IsAvailableCommand() {
		return false
	}

	_, hasGroup := cmd.Annotations["group"]
	return !hasGroup
}
