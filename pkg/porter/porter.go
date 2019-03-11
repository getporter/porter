//go:generate packr2

package porter

import (
	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/mixin/provider"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config
	*Templates
	MixinProvider
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	c := config.New()
	return &Porter{
		Config:        c,
		Templates: NewTemplates(),
		MixinProvider: mixinprovider.NewFileSystem(c),
	}
}
