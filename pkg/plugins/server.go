package plugins

import (
	"github.com/hashicorp/go-plugin"
)

// Serve a single named plugin.
func Serve(name string, p plugin.Plugin) {
	ServeMany(map[string]plugin.Plugin{name: p})
}

// Serve many plugins that the client will select by name.
func ServeMany(pluginMap map[string]plugin.Plugin) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         pluginMap,
	})
}
