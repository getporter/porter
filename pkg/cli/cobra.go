package cli

import (
	"fmt"
	"os"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"github.com/spf13/cobra"
)

// GetCalledCommand returns metadata about the command that was called.
func GetCalledCommand(cmd *cobra.Command) CalledCommand {
	var result CalledCommand

	// Determine what sub-command was called, such as porter installations list
	calledCommand, _, err := cmd.Find(os.Args[1:])
	if err != nil {
		result.CommandPath = cmd.Name()
	} else {
		result.CommandPath = calledCommand.CommandPath()
	}

	// Also figure out the full command called, with args/flags.
	result.FormattedCommand = fmt.Sprintf("%s %s", cmd.Name(), strings.Join(os.Args[1:], " "))

	// Detect if this command runs a plugin
	if IsPluginCommand(calledCommand) {
		result.IsPlugin = true
		// Get the first positional argument of the command called, such as plugin run PLUGIN_KEY
		firstArgIndex := strings.Count(calledCommand.CommandPath(), " ") + 1
		if len(os.Args) > firstArgIndex {
			result.PluginKey = os.Args[firstArgIndex]
		}
	}

	return result
}

type CalledCommand struct {
	// CommandPath of the command that was called, such as "porter installation show". The
	// arguments and flags are not included so that it can be used to tell which
	// command was called.
	CommandPath string

	// Cmd is the resolved cobra command.
	Cmd *cobra.Command

	// FormattedCommand is the fully-formatted command that was called, including arguments and flags.
	// This is useful for logging the requested command.
	FormattedCommand string

	// SkipConfig indicates that the command should not load the Porter configuration.
	// This is reserved for commands that should never fail, like help or version.
	SkipConfig bool

	// IsPlugin indicates if the command is running a plugin.
	IsPlugin bool

	// PluginKey is the key of the plugin that was requested by the command.
	PluginKey string
}

// ConfigureCommand applies configuration from the command to the Porter configuration.
func ConfigureCommand(rootCmd *cobra.Command, config *config.Config) CalledCommand {
	// Trace the command that called porter, e.g. porter installation show
	called := GetCalledCommand(rootCmd)

	// When running an internal plugin, switch how we log to be compatible
	// with the hashicorp go-plugin framework
	config.IsInternalPlugin = called.IsPlugin
	config.InternalPluginKey = called.PluginKey

	return called
}

// SkipConfigForCommand sets the command to skip loading Porter's configuration.
// This is useful for commands that should NEVER fail, like help or version.
func SkipConfigForCommand(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string, 1)
	}
	cmd.Annotations[AnnotationSkipConfig] = ""
}

// ShouldSkipConfig returns if the command has opted out of loading Porter's configuration.
func ShouldSkipConfig(cmd *cobra.Command) bool {
	if cmd.Name() == "help" {
		return true
	}

	_, skip := cmd.Annotations[AnnotationSkipConfig]
	return skip
}

// MarkCommandAsPlugin indicates that the command runs a plugin.
func MarkCommandAsPlugin(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string, 1)
	}
	cmd.Annotations[AnnotationIsPluginCommand] = ""
}

// IsPluginCommand returns if the command runs a plugin.
func IsPluginCommand(cmd *cobra.Command) bool {
	_, isPlugin := cmd.Annotations[AnnotationIsPluginCommand]
	return isPlugin
}

// SetCommandGroup indicates how the command should be grouped in the help text.
func SetCommandGroup(cmd *cobra.Command, group string) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string, 1)
	}
	cmd.Annotations[AnnotationGroup] = group
}
