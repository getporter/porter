package host

import (
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"github.com/cnabio/cnab-go/secrets/host"
)

const PluginKey = secrets.PluginInterface + ".porter.host"

var _ plugins.SecretsPlugin = Plugin{}

type Plugin struct {
	*host.SecretStore
}

func (p Plugin) Connect() error {
	return nil
}

func (p Plugin) Close() error {
	return nil
}

// NewPlugin creates an instance of the internal plugin secrets.porter.host
func NewPlugin() Plugin {
	return Plugin{SecretStore: &host.SecretStore{}}
}
