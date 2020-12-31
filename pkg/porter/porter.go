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
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/storage/pluginstore"
	"get.porter.sh/porter/pkg/templates"
	"github.com/cnabio/cnab-go/claim"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config

	Cache       cache.BundleCache
	Credentials credentials.CredentialProvider
	Parameters  parameters.ParameterProvider
	Claims      claim.Provider
	Registry    cnabtooci.RegistryProvider
	Templates   *templates.Templates
	Builder     BuildProvider
	Manifest    *manifest.Manifest
	Mixins      mixin.MixinProvider
	Plugins     plugins.PluginProvider
	CNAB        cnabprovider.CNABProvider
	Storage     storage.StorageProvider
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	c := config.New()
	return NewWithConfig(c)
}

func NewWithConfig(c *config.Config) *Porter {
	cache := cache.New(c)
	storagePlugin := pluginstore.NewStore(c)
	storageManager := storage.NewManager(c, storagePlugin)
	claimStorage := claims.NewClaimStorage(storageManager)
	credStorage := credentials.NewCredentialStorage(storageManager)
	paramStorage := parameters.NewParameterStorage(storageManager)
	return &Porter{
		Config:      c,
		Cache:       cache,
		Storage:     storageManager,
		Claims:      claimStorage,
		Credentials: credStorage,
		Parameters:  paramStorage,
		Registry:    cnabtooci.NewRegistry(c.Context),
		Templates:   templates.NewTemplates(),
		Builder:     buildprovider.NewDockerBuilder(c.Context),
		Mixins:      mixin.NewPackageManager(c),
		Plugins:     plugins.NewPackageManager(c),
		CNAB:        cnabprovider.NewRuntime(c, claimStorage, credStorage, paramStorage),
	}
}

// Init sets up the porter app and checks for common errors
// that would prevent Porter from running.
func (p *Porter) Init() error {
	// Do one-time checks before we run a porter command so that we aren't
	// checking for errors constantly in the rest of the code
	_, err := p.FindHomeDir()
	return err
}

func (p *Porter) LoadManifest() error {
	if p.Manifest != nil {
		return nil
	}
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
