package plugins

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// Serve a single named plugin.
func Serve(interfaceName string, pluginImplementation plugin.Plugin, protocolVersion int) {
	plugins := map[int]plugin.PluginSet{
		protocolVersion: {
			interfaceName: pluginImplementation,
		},
	}
	ServeMany(plugins)
}

// Serve many plugins that the client will select by named interface.
func ServeMany(pluginMap map[int]plugin.PluginSet) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "plugin",
		Output:     os.Stderr,
		Level:      hclog.Debug,
		JSONFormat: true,
	})

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig:  HandshakeConfig,
		VersionedPlugins: pluginMap,
		GRPCServer:       plugin.DefaultGRPCServer,
		Logger:           logger,
	})
}
