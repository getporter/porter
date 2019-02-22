package porter

import (
	"github.com/deislabs/porter/pkg/config"
	mixinprovider "github.com/deislabs/porter/pkg/mixin/provider"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config
	MixinProvider
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	c := config.New()
	return &Porter{
		Config:        c,
		MixinProvider: mixinprovider.NewFileSystem(c),
	}
}
