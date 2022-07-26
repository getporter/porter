package porter

import (
	"context"
	"errors"
	"strings"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/build/buildkit"
	"get.porter.sh/porter/pkg/cache"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/secrets"
	secretsplugin "get.porter.sh/porter/pkg/secrets/pluginstore"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/storage/migrations"
	storageplugin "get.porter.sh/porter/pkg/storage/pluginstore"
	"get.porter.sh/porter/pkg/templates"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/hashicorp/go-multierror"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config

	// builder is loaded dynamically when unset, this allows us to
	// use the configuration that is set after we create Porter,
	// or to switch it out for tests.
	builder build.Builder

	Cache         cache.BundleCache
	Credentials   storage.CredentialSetProvider
	Parameters    storage.ParameterSetProvider
	Sanitizer     *storage.Sanitizer
	Installations storage.InstallationProvider
	Registry      cnabtooci.RegistryProvider
	Templates     *templates.Templates
	Mixins        mixin.MixinProvider
	Plugins       plugins.PluginProvider
	CNAB          cnabprovider.CNABProvider
	Secrets       secrets.Store
	Storage       storage.Provider
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	c := config.New()
	storage := storage.NewPluginAdapter(storageplugin.NewStore(c))
	secretStorage := secrets.NewPluginAdapter(secretsplugin.NewStore(c))
	return NewFor(c, storage, secretStorage)
}

func NewFor(c *config.Config, store storage.Store, secretStorage secrets.Store) *Porter {
	cache := cache.New(c)

	storageManager := migrations.NewManager(c, store)
	installationStorage := storage.NewInstallationStore(storageManager)
	credStorage := storage.NewCredentialStore(storageManager, secretStorage)
	paramStorage := storage.NewParameterStore(storageManager, secretStorage)
	sanitizerService := storage.NewSanitizer(paramStorage, secretStorage)
	storageManager.Initialize(sanitizerService) // we have a bit of a dependency problem here that it would be great to figure out eventually

	return &Porter{
		Config:        c,
		Cache:         cache,
		Storage:       storageManager,
		Installations: installationStorage,
		Credentials:   credStorage,
		Parameters:    paramStorage,
		Secrets:       secretStorage,
		Registry:      cnabtooci.NewRegistry(c.Context),
		Templates:     templates.NewTemplates(c),
		Mixins:        mixin.NewPackageManager(c),
		Plugins:       plugins.NewPackageManager(c),
		CNAB:          cnabprovider.NewRuntime(c, installationStorage, credStorage, secretStorage, sanitizerService),
		Sanitizer:     sanitizerService,
	}
}

// Connect initializes Porter for use and must be called before other Porter methods.
// It is the responsibility of the caller to also call Close when done with Porter.
func (p *Porter) Connect(ctx context.Context) error {
	// Load the config file and replace any referenced secrets
	return p.Config.Load(ctx, func(innerCtx context.Context, secret string) (string, error) {
		value, err := p.Secrets.Resolve(innerCtx, "secret", secret)
		if err != nil {
			if strings.Contains(err.Error(), "invalid value source: secret") {
				return "", errors.New("No secret store account is configured")
			}
			return "", err
		}
		return value, nil
	})
}

// Close releases resources used by Porter before terminating the application.
func (p *Porter) Close() error {
	// Shutdown our plugins
	var bigErr *multierror.Error

	err := p.Secrets.Close()
	if err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	err = p.Storage.Close()
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
