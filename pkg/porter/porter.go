//go:generate packr2

package porter

import (
	"github.com/deislabs/porter/pkg/cache"
	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/config"
	mixinprovider "github.com/deislabs/porter/pkg/mixin/provider"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config
	Cache     cache.BundleCache
	Templates *Templates
	Mixins    MixinProvider
	CNAB      CNABProvider
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	c := config.New()
	cache := cache.New(c)
	return &Porter{
		Config:    c,
		Cache:     cache,
		Templates: NewTemplates(),
		Mixins:    mixinprovider.NewFileSystem(c),
		CNAB:      cnabprovider.NewDuffle(c),
	}
}
