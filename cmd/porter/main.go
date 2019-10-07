//go:generate packr2

package main

import (
	"os"

	"github.com/deislabs/porter/pkg/config/datastore"
	"github.com/deislabs/porter/pkg/porter"
	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/cobra"
)

var includeDocsCommand = false

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			p.Config.DataLoader = datastore.FromFlagsThenEnvVarsThenConfigFile(cmd)
			err := p.LoadData()
			if err != nil {
				return err
			}

			// Enable swapping out stdout/stderr for testing
			p.Out = cmd.OutOrStdout()
			p.Err = cmd.OutOrStderr()

			return nil
		},
		SilenceUsage: true,
	}

	cmd.PersistentFlags().BoolVar(&p.Debug, "debug", false, "Enable debug logging")

	cmd.AddCommand(buildVersionCommand(p))
	cmd.AddCommand(buildSchemaCommand(p))
	cmd.AddCommand(buildRunCommand(p))
	cmd.AddCommand(buildBundleCommands(p))
	cmd.AddCommand(buildInstanceCommands(p))
	cmd.AddCommand(buildMixinCommands(p))
	cmd.AddCommand(buildCredentialsCommands(p))

	for _, alias := range buildAliasCommands(p) {
		cmd.AddCommand(alias)
	}

	help := newHelptextBox()
	usage, _ := help.FindString("usage.txt")
	cmd.SetUsageTemplate(usage)
	cobra.AddTemplateFunc("ShouldShowGroupCommands", ShouldShowGroupCommands)
	cobra.AddTemplateFunc("ShouldShowGroupCommand", ShouldShowGroupCommand)
	cobra.AddTemplateFunc("ShouldShowUngroupedCommands", ShouldShowUngroupedCommands)
	cobra.AddTemplateFunc("ShouldShowUngroupedCommand", ShouldShowUngroupedCommand)

	if includeDocsCommand {
		cmd.AddCommand(buildDocsCommand(p))
	}

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
