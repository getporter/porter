package plugins

import (
	"github.com/hashicorp/go-plugin"
)

// Serve a single named plugin.
func Serve(interfaceName string, pluginImplementation plugin.Plugin) {
	ServeMany(map[string]plugin.Plugin{interfaceName: pluginImplementation})
}

// Serve many plugins that the client will select by named interface.
func ServeMany(pluginMap map[string]plugin.Plugin) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         pluginMap,
	})
}
