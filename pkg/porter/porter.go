//go:generate packr2

package porter

import (
	buildprovider "github.com/deislabs/porter/pkg/build/provider"
	"github.com/deislabs/porter/pkg/cache"
	cnabtooci "github.com/deislabs/porter/pkg/cnab/cnab-to-oci"
	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/mixin"
	mixinprovider "github.com/deislabs/porter/pkg/mixin/provider"
	"github.com/deislabs/porter/pkg/templates"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config
	Cache     cache.BundleCache
	Registry  Registry
	Templates *templates.Templates
	Builder   BuildProvider
	Mixins    mixin.MixinProvider
	CNAB      CNABProvider
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	c := config.New()
	cache := cache.New(c)
	return &Porter{
		Config:    c,
		Cache:     cache,
		Registry:  cnabtooci.NewRegistry(c.Context),
		Templates: templates.NewTemplates(),
		Builder:   buildprovider.NewDockerBuilder(c),
		Mixins:    mixinprovider.NewFileSystem(c),
		CNAB:      cnabprovider.NewRuntime(c),
	}
}
