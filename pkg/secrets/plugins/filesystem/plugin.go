package filesystem

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/pluginstore"
	"github.com/hashicorp/go-plugin"
)

const PluginKey = plugins.PluginInterface + ".porter.filesystem"

var _ plugins.SecretsProtocol = &Plugin{}

// Plugin is the plugin wrapper for accessing secrets from a local filesystem.
type Plugin struct {
	secrets.Store
}

func NewPlugin(c *config.Config, rawCfg interface{}) plugin.Plugin {
	impl := NewStore(c)
	return pluginstore.NewPlugin(c.Context, impl)
}
