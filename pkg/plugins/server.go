package plugins

import (
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
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig:  HandshakeConfig,
		VersionedPlugins: pluginMap,
		GRPCServer:       plugin.DefaultGRPCServer,
	})
}
