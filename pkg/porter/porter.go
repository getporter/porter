//go:generate packr2

package porter

import (
	buildprovider "get.porter.sh/porter/pkg/build/provider"
	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/claims"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	mixinprovider "get.porter.sh/porter/pkg/mixin/provider"
	"get.porter.sh/porter/pkg/storage/pluginstore"
	"get.porter.sh/porter/pkg/templates"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config

	Cache       cache.BundleCache
	Credentials credentials.CredentialProvider
	Claims      claims.ClaimProvider
	Registry    Registry
	Templates   *templates.Templates
	Builder     BuildProvider
	Manifest    *manifest.Manifest
	Mixins      mixin.MixinProvider
	CNAB        CNABProvider
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	c := config.New()
	cache := cache.New(c)
	storagePlugin := pluginstore.NewStore(c)
	claimStorage := claims.NewClaimStorage(c, storagePlugin)
	credStorage := credentials.NewCredentialStorage(c, storagePlugin)
	return &Porter{
		Config:      c,
		Cache:       cache,
		Claims:      claimStorage,
		Credentials: credStorage,
		Registry:    cnabtooci.NewRegistry(c.Context),
		Templates:   templates.NewTemplates(),
		Builder:     buildprovider.NewDockerBuilder(c.Context),
		Mixins:      mixinprovider.NewFileSystem(c),
		CNAB:        cnabprovider.NewRuntime(c, claimStorage, credStorage),
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
