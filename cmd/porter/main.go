package main

import (
	"context"
	_ "embed"

	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var includeDocsCommand = false

//go:embed helptext/usage.txt
var usageText string

func main() {
	ctx := context.Background()
	app := porter.New()
	rootCmd := buildRootCommandFrom(app)

	cli.Main(ctx, rootCmd, app)
}

func buildRootCommand() *cobra.Command {
	return buildRootCommandFrom(porter.New())
}

func buildRootCommandFrom(p *porter.Porter) *cobra.Command {
	var printVersion bool

	cmd := &cobra.Command{
		Use: "porter",
		Short: `With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://getporter.org/quickstart to learn how to use Porter.
`,
		Example: `  porter create
  porter build
  porter install
  porter uninstall`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Enable swapping out stdout/stderr for testing
			p.Out = cmd.OutOrStdout()
			p.Err = cmd.OutOrStderr()

			if cli.ShouldSkipConfig(cmd) {
				return nil
			}

			// Reload configuration with the now parsed cli flags
			p.DataLoader = cli.LoadHierarchicalConfig(cmd)
			err := p.Connect(cmd.Context())
			if err != nil {
				return err
			}

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
		SilenceUsage:  true,
		SilenceErrors: true, // Errors are printed by main
	}

	cli.SkipConfigForCommand(cmd)

	// These flags are available for every command
	globalFlags := cmd.PersistentFlags()
	globalFlags.BoolVar(&p.Debug, "debug", false, "Enable debug logging")
	globalFlags.BoolVar(&p.DebugPlugins, "debug-plugins", false, "Enable plugin debug logging")
	globalFlags.StringSliceVar(&p.Data.ExperimentalFlags, "experimental", nil, "Comma separated list of experimental features to enable. See https://getporter.org/configuration/#experimental-feature-flags for available feature flags.")

	// Flags for just the porter command only, does not apply to sub-commands
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
	cmd.AddCommand(buildCompletionCommand(p))

	for _, alias := range buildAliasCommands(p) {
		cmd.AddCommand(alias)
	}

	cmd.SetUsageTemplate(usageText)
	cobra.AddTemplateFunc("ShouldShowGroupCommands", ShouldShowGroupCommands)
	cobra.AddTemplateFunc("ShouldShowGroupCommand", ShouldShowGroupCommand)
	cobra.AddTemplateFunc("ShouldShowUngroupedCommands", ShouldShowUngroupedCommands)
	cobra.AddTemplateFunc("ShouldShowUngroupedCommand", ShouldShowUngroupedCommand)

	if includeDocsCommand {
		cmd.AddCommand(buildDocsCommand(p))
	}

	return cmd
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
	if cmd.Annotations[cli.AnnotationGroup] == group {
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

	_, hasGroup := cmd.Annotations[cli.AnnotationGroup]
	return !hasGroup
}

func addBundlePullFlags(f *pflag.FlagSet, opts *porter.BundlePullOptions) {
	addReferenceFlag(f, opts)
	addInsecureRegistryFlag(f, opts)
	addForcePullFlag(f, opts)
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
