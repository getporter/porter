package filesystem

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/crudstore"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

const PluginKey = crudstore.PluginInterface + ".porter.filesystem"

var _ crud.Store = &Plugin{}

// Plugin is the plugin wrapper for the local filesystem storage.
type Plugin struct {
	crud.Store
}

func NewPlugin(c config.Config) plugin.Plugin {
	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       PluginKey,
		Output:     c.Err,
		Level:      hclog.Debug,
		JSONFormat: true,
	})

	return &crudstore.Plugin{
		Impl: &Plugin{
			Store: NewStore(c, logger),
		},
	}
}
