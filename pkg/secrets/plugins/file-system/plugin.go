package filesystem

import (
	"os"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"github.com/hashicorp/go-hclog"
)

const PluginKey = secrets.PluginInterface + ".porter.file-system"

var _ plugins.SecretsPlugin = &Plugin{}

// Plugin is the plugin wrapper for accessing secrets from Azure Key Vault.
type Plugin struct {
	secrets.Store
}

func NewPlugin() Plugin {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       PluginKey,
		Output:     os.Stderr,
		Level:      hclog.Debug,
		JSONFormat: true,
	})

	cfg := Config{}

	return Plugin{
		Store: NewStore(cfg, logger),
	}

}

func (p Plugin) Connect() error {
	return p.Store.Connect()
}

func (p Plugin) Close() error {
	return p.Store.Close()
}
