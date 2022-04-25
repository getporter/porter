package filesystem

import (
	"os"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"github.com/carolynvs/aferox"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/afero"
)

const PluginKey = secrets.PluginInterface + ".porter.filesystem"

var _ plugins.SecretsPlugin = &Plugin{}

// Plugin is the plugin wrapper for accessing secrets from a local filesystem.
type Plugin struct {
	secrets.Store
}

func NewPlugin(cfg Config) (Plugin, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       PluginKey,
		Output:     os.Stderr,
		Level:      hclog.Info,
		JSONFormat: true,
	})
	_, err := cfg.SetSecretDir()
	if err != nil {
		return Plugin{}, err
	}

	return Plugin{
		Store: NewStore(cfg, logger, aferox.NewAferox(cfg.secretDir, afero.NewOsFs())),
	}, nil

}

func (p Plugin) Connect() error {
	return p.Store.Connect()
}

func (p Plugin) Close() error {
	return p.Store.Close()
}
