package noop

import (
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/pluginstore"
	"github.com/hashicorp/go-plugin"
)

const PluginKey = plugins.PluginInterface + ".porter.noop"

var _ plugins.StorageProtocol = Plugin{}

type Plugin struct {
	*Store
}

func NewPlugin(c *portercontext.Context) (plugin.Plugin, error) {
	impl := NewStore()
	return pluginstore.NewPlugin(c, impl), nil
}
