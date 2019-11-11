package secrets

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// PluginInterface for the secrets. This first part of the
// three-part plugin key is only seen/used by the plugins when the host is
// communicating with the plugin and is not exposed to users.
const PluginInterface = "secrets"

var _ plugin.Plugin = &Plugin{}

// Plugin is a generic type of plugin for working with any implementation of a secret store.
type Plugin struct {
	Impl Store
}

func (p *Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &Server{Impl: p.Impl}, nil
}

func (Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &Client{client: c}, nil
}
