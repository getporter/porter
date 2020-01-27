//go:generate packr2

package porter

import (
	buildprovider "get.porter.sh/porter/pkg/build/provider"
	"get.porter.sh/porter/pkg/cache"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	instancestorage "get.porter.sh/porter/pkg/instance-storage"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	mixinprovider "get.porter.sh/porter/pkg/mixin/provider"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/templates"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config

	Cache           cache.BundleCache
	InstanceStorage instancestorage.StorageProvider
	Registry        Registry
	Templates       *templates.Templates
	Builder         BuildProvider
	Manifest        *manifest.Manifest
	Mixins          mixin.MixinProvider
	Plugins         plugins.PluginProvider
	CNAB            CNABProvider
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	c := config.New()
	cache := cache.New(c)
	instanceStorage := instancestorage.NewPluggableInstanceStorage(c)
	return &Porter{
		Config:          c,
		Cache:           cache,
		InstanceStorage: instanceStorage,
		Registry:        cnabtooci.NewRegistry(c.Context),
		Templates:       templates.NewTemplates(),
		Builder:         buildprovider.NewDockerBuilder(c.Context),
		Mixins:          mixinprovider.NewFileSystem(c),
		CNAB:            cnabprovider.NewRuntime(c, instanceStorage),
	}
}

func (p *Porter) LoadManifest() error {
	return p.LoadManifestFrom(config.Name)
}

func (p *Porter) LoadManifestFrom(file string) error {
	m, err := manifest.LoadManifestFrom(p.Context, file)
	if err != nil {
		return err
	}
	p.Manifest = m
	return nil
}
