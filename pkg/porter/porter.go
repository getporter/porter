package porter

import (
	"context"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/build/buildkit"
	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/claims"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/secrets"
	secretsplugin "get.porter.sh/porter/pkg/secrets/pluginstore"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/storage/migrations"
	"get.porter.sh/porter/pkg/storage/pluginstore"
	"get.porter.sh/porter/pkg/templates"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config

	// builder is loaded dynamically when unset, this allows us to
	// use the configuration that is set after we create Porter,
	// or to switch it out for tests.
	builder build.Builder

	Cache       cache.BundleCache
	Credentials credentials.Provider
	Parameters  parameters.Provider
	Claims      claims.Provider
	Registry    cnabtooci.RegistryProvider
	Templates   *templates.Templates
	Mixins      mixin.MixinProvider
	Plugins     plugins.PluginProvider
	CNAB        cnabprovider.CNABProvider
	Secrets     secrets.Store
	Storage     storage.Provider
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	c := config.New()
	storagePlugin := pluginstore.NewStore(c)
	storage := storage.NewPluginAdapter(storagePlugin)
	return NewFor(c, storage)
}

func NewFor(c *config.Config, store storage.Store) *Porter {
	cache := cache.New(c)

	storageManager := migrations.NewManager(c, store)
	secretStorage := secrets.NewPluginAdapter(secretsplugin.NewStore(c))
	claimStorage := claims.NewClaimStore(storageManager)
	credStorage := credentials.NewCredentialStore(storageManager, secretStorage)
	paramStorage := parameters.NewParameterStore(storageManager, secretStorage)
	return &Porter{
		Config:      c,
		Cache:       cache,
		Storage:     storageManager,
		Claims:      claimStorage,
		Credentials: credStorage,
		Parameters:  paramStorage,
		Secrets:     secretStorage,
		Registry:    cnabtooci.NewRegistry(c.Context),
		Templates:   templates.NewTemplates(c),
		Mixins:      mixin.NewPackageManager(c),
		Plugins:     plugins.NewPackageManager(c),
		CNAB:        cnabprovider.NewRuntime(c, claimStorage, credStorage),
	}
}

func (p *Porter) Connect(ctx context.Context) error {
	// Load the config file and replace any referenced secrets
	return p.Config.Load(ctx, func(secret string) (string, error) {
		value, err := p.Secrets.Resolve(ctx, "secret", secret)
		if err != nil {
			if strings.Contains(err.Error(), "invalid value source: secret") {
				return "", errors.New("No secret store account is configured")
			}
		}
		return value, nil
	})
}

// Close releases resources used by Porter before terminating the application.
func (p *Porter) Close(ctx context.Context) error {
	if p.Debug {
		fmt.Fprintln(p.Err, "Closing plugins")
	}

	// Shutdown our plugins
	var bigErr *multierror.Error

	err := p.Secrets.Close(ctx)
	if err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	err = p.Storage.Close(ctx)
	if err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	err = p.Config.Close()
	if err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	return bigErr.ErrorOrNil()
}

// GetBuilder creates a Builder based on the current configuration.
func (p *Porter) GetBuilder(ctx context.Context) build.Builder {
	log := tracing.LoggerFromContext(ctx)

	if p.builder == nil {
		driver := p.GetBuildDriver()
		switch driver {
		case config.BuildDriverBuildkit:
			// supported, yay!
		case config.BuildDriverDocker:
			log.Warn("The docker build driver is no longer supported. Using buildkit instead.")
		default:
			log.Warnf("Unsupported build driver: %s. Using buildkit instead.", driver)
		}
		p.builder = buildkit.NewBuilder(p.Config)
	}
	return p.builder
}
