package pluginbuilder

import (
	"context"
	"strings"

	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
)

var _ cli.PorterApp = &PorterPlugin{}

// PorterPlugin is a reusable helper that provides most of the basic plugin
// functionality for a Porter plugin.
type PorterPlugin struct {
	// Opts customizes the default implementation of plugin.
	opts PluginOptions

	// porterConfig provides Porter's configuration. This is helpful for retrieving
	// the PORTER_HOME directory, and using the portercontext.Context field to make
	// the plugin more easily testable.
	porterConfig *config.Config

	// pluginConfig holds any plugin-specific configuration defined in the Porter
	// configuration file.
	pluginConfig interface{}
}

// Connect prepares the plugin to run.
func (p *PorterPlugin) Connect(ctx context.Context) error {
	// Load the Porter configuration
	return p.porterConfig.Load(ctx, nil)
}

// Close the plugin and release any resources held.
func (p *PorterPlugin) Close() error {
	// Release resources held by the Porter configuration
	return p.porterConfig.Close()
}

// GetConfig returns the plugin's Porter configuration.
func (p *PorterPlugin) GetConfig() *config.Config {
	return p.porterConfig
}

// Name returns the plugin name.
func (p *PorterPlugin) Name() string {
	return p.opts.Name
}

// PluginOptions customizes the default plugin implementation of your custom
// plugin.
type PluginOptions struct {
	// Name of the plugin.
	Name string

	// DefaultConfig contains the default configuration data structure into which
	// the plugin's configuration will be placed when the plugin is run.
	// Defaults to a map[string]interface{}.
	DefaultConfig interface{}

	// RegisteredPlugins is a lookup from a fully-qualified 3-part plugin key
	// to the information necessary to run the plugin.
	RegisteredPlugins map[string]plugins.PluginRegistration

	// Version is the semantic version for this build of the plugin.
	Version string

	// Commit is the git commit hash for this build of the plugin.
	Commit string
}

// NewPlugin creates a PorterPlugin and customizes the implementation using the
// specified PluginOptions.
func NewPlugin(opts PluginOptions) *PorterPlugin {
	return &PorterPlugin{
		opts:         opts,
		porterConfig: config.New(),
	}
}

// ListSupportedKeys prints a human-readable list of the plugin keys supported by
// the plugin.
func (opts *PluginOptions) ListSupportedKeys() string {
	keys := make([]string, 0, len(opts.RegisteredPlugins))
	for key := range opts.RegisteredPlugins {
		keys = append(keys, key)
	}
	return strings.Join(keys, ", ")
}
