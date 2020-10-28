package host

import (
	"get.porter.sh/porter/pkg/secrets"
	cnabsecrets "github.com/cnabio/cnab-go/secrets"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/hashicorp/go-plugin"
)

const PluginKey = secrets.PluginInterface + ".porter.host"

var _ cnabsecrets.Store = &Plugin{}

// Plugin is the plugin wrapper for the local host secrets.
type Plugin struct {
	cnabsecrets.Store
}

func NewPlugin() plugin.Plugin {
	return &secrets.Plugin{
		Impl: &Plugin{
			Store: &host.SecretStore{},
		},
	}
}
