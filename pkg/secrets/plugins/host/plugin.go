package host

import (
	"context"

	"github.com/hashicorp/go-plugin"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"github.com/cnabio/cnab-go/secrets/host"
)

const PluginKey = secrets.PluginInterface + ".porter.host"

var _ plugins.SecretsPlugin = Plugin{}

type Plugin struct {
	*host.SecretStore
}

func NewPlugin() plugin.Plugin {
	return &secrets.Plugin{
		Impl: &Plugin{
			SecretStore: &host.SecretStore{},
		},
	}
}

func (p Plugin) Connect(context.Context) error {
	return nil
}

func (p Plugin) Close(context.Context) error {
	return nil
}
