//go:generate packr2

package main

import (
	"os"

	"get.porter.sh/porter/pkg/config/datastore"
	"get.porter.sh/porter/pkg/porter"
	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	var printVersion bool

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
		RunE: func(cmd *cobra.Command, args []string) error {
			if printVersion {
				versionCmd := buildVersionCommand(p)
				err := versionCmd.PreRunE(cmd, args)
				if err != nil {
					return err
				}
				return versionCmd.RunE(cmd, args)
			}
			return cmd.Help()
		},
		SilenceUsage: true,
	}

	cmd.PersistentFlags().BoolVar(&p.Debug, "debug", false, "Enable debug logging")
	cmd.PersistentFlags().BoolVar(&p.DebugPlugins, "debug-plugins", false, "Enable plugin debug logging")

	cmd.Flags().BoolVarP(&printVersion, "version", "v", false, "Print the application version")

	cmd.AddCommand(buildVersionCommand(p))
	cmd.AddCommand(buildSchemaCommand(p))
	cmd.AddCommand(buildStorageCommand(p))
	cmd.AddCommand(buildRunCommand(p))
	cmd.AddCommand(buildBundleCommands(p))
	cmd.AddCommand(buildInstallationCommands(p))
	cmd.AddCommand(buildMixinCommands(p))
	cmd.AddCommand(buildPluginsCommands(p))
	cmd.AddCommand(buildCredentialsCommands(p))
	cmd.AddCommand(buildParametersCommands(p))

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
	return packr.New("get.porter.sh/porter/cmd/porter/helptext", "./helptext")
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

func addBundlePullFlags(f *pflag.FlagSet, opts *porter.BundlePullOptions) {
	addDeprecatedTagFlag(f, opts)
	addReferenceFlag(f, opts)
	addInsecureRegistryFlag(f, opts)
	addForcePullFlag(f, opts)
}

func addDeprecatedTagFlag(f *pflag.FlagSet, opts *porter.BundlePullOptions) {
	f.StringVar(&opts.Tag, "tag", "", "")
	f.MarkDeprecated("tag", "use --reference to declare a full bundle reference")
}

func addReferenceFlag(f *pflag.FlagSet, opts *porter.BundlePullOptions) {
	f.StringVarP(&opts.Reference, "reference", "r", "",
		"Use a bundle in an OCI registry specified by the given reference.")
}

func addInsecureRegistryFlag(f *pflag.FlagSet, opts *porter.BundlePullOptions) {
	f.BoolVar(&opts.InsecureRegistry, "insecure-registry", false,
		"Don't require TLS for the registry")
}

func addForcePullFlag(f *pflag.FlagSet, opts *porter.BundlePullOptions) {
	f.BoolVar(&opts.Force, "force", false,
		"Force a fresh pull of the bundle")
}
