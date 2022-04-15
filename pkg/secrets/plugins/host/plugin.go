package host

import (
	"context"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"github.com/cnabio/cnab-go/secrets/host"
)

const PluginKey = secrets.PluginInterface + ".porter.host"

var _ plugins.SecretsPlugin = Plugin{}

type Plugin struct {
	*host.SecretStore
}

func (p Plugin) Connect(context.Context) error {
	return nil
}

func (p Plugin) Close(context.Context) error {
	return nil
}

// NewPlugin creates an instance of the internal plugin secrets.porter.host
func NewPlugin(ctx context.Context) Plugin {
	return Plugin{SecretStore: &host.SecretStore{}}
}
