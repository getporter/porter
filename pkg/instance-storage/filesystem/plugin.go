package filesystem

import (
	"github.com/deislabs/cnab-go/utils/crud"
	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/instance-storage/claimstore"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

const PluginKey = claimstore.PluginKey + ".filesystem"

var _ crud.Store = &Plugin{}

// A sad hack because crud.Store has a method called Store which prevents us from embedding it as a field
type CrudStore = crud.Store

// Plugin is the plugin wrapper for the local filesystem storage for claims.
type Plugin struct {
	CrudStore
}

func NewPlugin(c config.Config) plugin.Plugin {
	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   PluginKey,
		Output: c.Err,
		Level:  hclog.Warn,
	})

	return &claimstore.Plugin{
		Impl: &Plugin{
			CrudStore: NewStore(c, logger),
		},
	}
}
