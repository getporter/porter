package host

import (
	"fmt"

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

func (p Plugin) Create(keyName string, keyValue string, value string) error {
	return fmt.Errorf("the %s plugin does not support saving secrets. Please configure a secret plugin", PluginKey)
}

// NewPlugin creates an instance of the internal plugin secrets.porter.host
func NewPlugin() Plugin {
	return Plugin{SecretStore: &host.SecretStore{}}
}
