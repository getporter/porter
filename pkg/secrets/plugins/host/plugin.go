package host

import (
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/pluginstore"
	"github.com/hashicorp/go-plugin"
)

const PluginKey = plugins.PluginInterface + ".porter.host"

var _ plugins.SecretsProtocol = Plugin{}

type Plugin struct {
	Store
}

func NewPlugin(c *portercontext.Context) plugin.Plugin {
	store := NewStore()
	return pluginstore.NewPlugin(c, store)
}
