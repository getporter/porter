package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/porter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel/attribute"
)

var includeDocsCommand = false

//go:embed helptext/usage.txt
var usageText string

// Indicates that config should not be loaded for this command.
// This is used for commands like help and version which should never
// fail, even with porter is misconfigured.
const skipConfig string = "skipConfig"

func main() {
	run := func() int {
		ctx := context.Background()
		p := porter.New()

		rootCmd := buildRootCommandFrom(p)

		// Trace the command that called porter, e.g. porter installation show
		cmd, commandName, formattedCommand := getCalledCommand(rootCmd)

		// When running an internal plugin, switch how we log to be compatible
		// with the hashicorp go-plugin framework
		if commandName == "porter plugins run" {
			p.IsInternalPlugin = true
			if len(os.Args) > 3 {
				p.InternalPluginKey = os.Args[3]
			}
		}

		// Only run init logic that could fail for commands that
		// really need it, skip it for commands that should NEVER
		// fail.
		if !shouldSkipConfig(cmd) {
			if err := p.Connect(ctx); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		}

		ctx, log := p.StartRootSpan(ctx, commandName, attribute.String("command", formattedCommand))
		defer func() {
			// Capture panics and trace them
			if panicErr := recover(); panicErr != nil {
				log.Error(errors.New(fmt.Sprintf("%s", panicErr)),
					attribute.Bool("panic", true),
					attribute.String("stackTrace", string(debug.Stack())))
				log.EndSpan()
				p.Close()
				os.Exit(1)
			} else {
				log.EndSpan()
				log.Close()
				p.Close()
			}
		}()

		if err := rootCmd.ExecuteContext(ctx); err != nil {
			// Ideally we log all errors in the span that generated it,
			// but as a failsafe, always log the error at the root span as well
			log.Error(err)
			return 1
		}
		return 0
	}

	// Wrapping the main run logic in a function because os.Exit will not
	// execute defer statements
	os.Exit(run())
}

func shouldSkipConfig(cmd *cobra.Command) bool {
	if cmd.Name() == "help" {
		return true
	}

	_, skip := cmd.Annotations[skipConfig]
	return skip
}

// Returns the porter command called, e.g. porter installation list
// and also the fully formatted command as passed with arguments/flags.
func getCalledCommand(cmd *cobra.Command) (*cobra.Command, string, string) {
	// Ask cobra what sub-command was called, and walk up the tree to get the full command called.
	var cmdChain []string
	calledCommand, _, err := cmd.Find(os.Args[1:])
	if err != nil {
		cmdChain = append(cmdChain, "porter")
	} else {
		cmd := calledCommand
		for cmd != nil {
			cmdChain = append(cmdChain, cmd.Name())
			cmd = cmd.Parent()
		}
	}
	// reverse the command from [list installations porter] to porter installation list
	var calledCommandBuilder strings.Builder
	for i := len(cmdChain); i > 0; i-- {
		calledCommandBuilder.WriteString(cmdChain[i-1])
		calledCommandBuilder.WriteString(" ")
	}
	calledCommandStr := calledCommandBuilder.String()[0 : calledCommandBuilder.Len()-1]

	// Also figure out the full command called, with args/flags.
	formattedCommand := fmt.Sprintf("porter %s", strings.Join(os.Args[1:], " "))

	return calledCommand, calledCommandStr, formattedCommand
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

Try our QuickStart https://porter.sh/quickstart to learn how to use Porter.
`,
		Example: `  porter create
  porter build
  porter install
  porter uninstall`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Enable swapping out stdout/stderr for testing
			p.Out = cmd.OutOrStdout()
			p.Err = cmd.OutOrStderr()

			if shouldSkipConfig(cmd) {
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

	cmd.Annotations = map[string]string{
		skipConfig: "",
	}

	// These flags are available for every command
	globalFlags := cmd.PersistentFlags()
	globalFlags.BoolVar(&p.Debug, "debug", false, "Enable debug logging")
	globalFlags.BoolVar(&p.DebugPlugins, "debug-plugins", false, "Enable plugin debug logging")
	globalFlags.StringSliceVar(&p.Data.ExperimentalFlags, "experimental", nil, "Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.")

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
