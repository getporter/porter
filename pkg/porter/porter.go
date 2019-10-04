//go:generate packr2

package porter

import (
	buildprovider "github.com/deislabs/porter/pkg/build/provider"
	"github.com/deislabs/porter/pkg/cache"
	cnabtooci "github.com/deislabs/porter/pkg/cnab/cnab-to-oci"
	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/config"
	instancestorage "github.com/deislabs/porter/pkg/instance-storage"
	instancestorageprovider "github.com/deislabs/porter/pkg/instance-storage/provider"
	"github.com/deislabs/porter/pkg/manifest"
	"github.com/deislabs/porter/pkg/mixin"
	mixinprovider "github.com/deislabs/porter/pkg/mixin/provider"
	"github.com/deislabs/porter/pkg/templates"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config

	Cache           cache.BundleCache
	InstanceStorage instancestorage.Provider
	Registry        Registry
	Templates       *templates.Templates
	Builder         BuildProvider
	Manifest        *manifest.Manifest
	Mixins          mixin.MixinProvider
	CNAB            CNABProvider
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	c := config.New()
	cache := cache.New(c)
	instanceStorage := instancestorageprovider.NewPluginDelegator(c)
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
