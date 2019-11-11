package host

import (
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage/crudstore"
	"github.com/hashicorp/go-plugin"
)

const PluginKey = crudstore.PluginInterface + ".porter.host"

var _ secrets.Store = &Plugin{}

// Plugin is the plugin wrapper for the local host secrets.
type Plugin struct {
	secrets.Store
}

func NewPlugin() plugin.Plugin {
	return &secrets.Plugin{
		Impl: &Plugin{
			Store: NewStore(),
		},
	}
}
