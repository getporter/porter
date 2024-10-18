package porter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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
	"get.porter.sh/porter/pkg/signing"
	signingplugin "get.porter.sh/porter/pkg/signing/pluginstore"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/storage/migrations"
	storageplugin "get.porter.sh/porter/pkg/storage/pluginstore"
	"get.porter.sh/porter/pkg/storage/sql"
	"get.porter.sh/porter/pkg/storage/sql/migrate"
	"get.porter.sh/porter/pkg/templates"
	"get.porter.sh/porter/pkg/tracing"
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
	Signer        signing.Signer

	onClose []func() error

	// Deprecated: Use the individual storage providers in the Porter struct instead
	// This is only here for backwards compatibility where MongoDB was the only storage provider.
	Storage storage.Provider
}

// Options for configuring a new Porter client passed to NewWith.
type Options struct {
	Config        *config.Config                // Optional. Defaults to a config.New.
	Secrets       secrets.Store                 // Optional. Defaults to a secrets.NewPluginAdapter(secretsplugin.NewStore).
	Signer        signing.Signer                // Optional. Defaults to a signing.NewPluginAdapter(signingplugin.NewSigner).
	Installations storage.InstallationProvider  // Optional. Defaults to a storage.NewInstallationStore for MongoDB or storage.NewInstallationStoreSQL for SQL.
	Credentials   storage.CredentialSetProvider // Optional. Defaults to a storage.NewCredentialStore for MongoDB or storage.NewCredentialStoreSQL for SQL.
	Parameters    storage.ParameterSetProvider  // Optional. Defaults to a storage.NewParameterStore for MongoDB or storage.NewParameterStoreSQL for SQL.
	Sanitizer     *storage.Sanitizer            // Optional. Defaults to a storage.NewSanitizer for MongoDB or storage.NewSanitizer for SQL.
}

// NewWith creates a new Porter client with useful defaults that can be overridden with the provided options.
func NewWith(opts Options) *Porter {
	return newWith(opts.Config, opts.Installations, opts.Credentials, opts.Parameters, opts.Secrets, opts.Signer, opts.Sanitizer, nil)
}

// New porter client, initialized with useful defaults.
// Use NewWith to override the defaults.
func New() *Porter {
	// differ full initialization to Connect
	return newWith(nil, nil, nil, nil, nil, nil, nil, nil)
}

// NewFor creates a new Porter client with the provided configuration and storage backend.
//
// Deprecated: Use NewWith instead. NewFor does not support newer sql storage backends.
func NewFor(
	c *config.Config,
	store storage.Store,
	secretStorage secrets.Store,
	signer signing.Signer,
) *Porter {
	storageManager := migrations.NewManager(c, store)
	installationStorage := storage.NewInstallationStore(storageManager)
	credStorage := storage.NewCredentialStore(storageManager, secretStorage)
	paramStorage := storage.NewParameterStore(storageManager, secretStorage)
	sanitizerService := storage.NewSanitizer(paramStorage, secretStorage)

	storageManager.Initialize(sanitizerService) // we have a bit of a dependency problem here that it would be great to figure out eventually

	return newWith(c, installationStorage, credStorage, paramStorage, secretStorage, signer, sanitizerService, storageManager)
}

func newWith(
	c *config.Config,
	installationStorage storage.InstallationProvider,
	credStorage storage.CredentialSetProvider,
	paramStorage storage.ParameterSetProvider,
	secretStorage secrets.Store,
	signer signing.Signer,
	sanitizerService *storage.Sanitizer,
	storageManager storage.Provider,
) *Porter {
	if c == nil {
		c = config.New()
	}
	if secretStorage == nil {
		secretStorage = secrets.NewPluginAdapter(secretsplugin.NewStore(c))
	}
	if signer == nil {
		signer = signing.NewPluginAdapter(signingplugin.NewSigner(c))
	}

	var cnab cnabprovider.CNABProvider
	if installationStorage != nil && credStorage != nil && paramStorage != nil && secretStorage != nil && sanitizerService != nil {
		cnab = cnabprovider.NewRuntime(c, installationStorage, credStorage, paramStorage, secretStorage, sanitizerService)
	}

	return &Porter{
		Config:        c,
		Cache:         cache.New(c),
		Storage:       storageManager,
		Installations: installationStorage,
		Credentials:   credStorage,
		Parameters:    paramStorage,
		Secrets:       secretStorage,
		Registry:      cnabtooci.NewRegistry(c.Context),
		Templates:     templates.NewTemplates(c),
		Mixins:        mixin.NewPackageManager(c),
		Plugins:       plugins.NewPackageManager(c),
		CNAB:          cnab,
		Sanitizer:     sanitizerService,
		Signer:        signer,
	}
}

// Used to warn just a single time when Porter starts up.
// Connect is called more than once, and this helps us validate certain things, like build flags, a single time only.
var initWarnings sync.Once

// Connect initializes Porter for use and must be called before other Porter methods.
// It is the responsibility of the caller to also call Close when done with Porter.
func (p *Porter) Connect(ctx context.Context) (context.Context, error) {
	initWarnings.Do(func() {
		// Check if this is a special dev build that will trace sensitive data and strongly warn people
		if tracing.IsTraceSensitiveAttributesEnabled() {
			fmt.Fprintln(p.Err, "🚨 WARNING! This is a custom developer build of Porter with the traceSensitiveAttributes build flag set. "+
				"Porter will include sensitive data, such as parameters and credentials, in the telemetry trace data. "+
				"This build flag should only be used for local development only. "+
				"If you didn't intend to use a custom build of Porter with this flag enabled, reinstall Porter using the official builds from https://porter.sh/install.")
		}
	})

	init := func(ctx context.Context) error {
		if p.Installations != nil && p.Credentials != nil && p.Parameters != nil && p.Sanitizer != nil {
			return nil // already initialized
		}

		storagePlugin, ok := sql.IsPostgresStorage(p.Config)
		if !ok {
			store := storage.NewPluginAdapter(storageplugin.NewStore(p.Config))
			mgr := migrations.NewManager(p.Config, store)

			p.Storage = mgr
			p.Installations = storage.NewInstallationStore(p.Storage)
			p.Credentials = storage.NewCredentialStore(p.Storage, p.Secrets)
			p.Parameters = storage.NewParameterStore(p.Storage, p.Secrets)
			p.Sanitizer = storage.NewSanitizer(p.Parameters, p.Secrets)

			mgr.Initialize(p.Sanitizer) // we have a bit of a dependency problem here that it would be great to figure out eventually
			return nil
		}

		// Initialize the SQL storage backend
		pc, err := sql.UnmarshalPluginConfig(storagePlugin.GetConfig())
		if err != nil {
			return fmt.Errorf("could not unmarshal plugin config: %s", err)
		}
		if pc.URL == "" {
			return errors.New("no URL provided in plugin config")
		}
		db, err := gorm.Open(postgres.Open(pc.URL))
		if err != nil {
			return fmt.Errorf("could not open database: %s", err)
		}
		closeFn := func() error {
			if sqlDB, err := db.DB(); err == nil {
				return sqlDB.Close()
			}
			return nil
		}
		if err = migrate.MigrateDB(ctx, db); err != nil {
			_ = closeFn()
			return err
		}

		p.onClose = append(p.onClose, closeFn)

		if p.Installations == nil {
			p.Installations = storage.NewInstallationStoreSQL(db)
		}
		if p.Credentials == nil {
			p.Credentials = storage.NewCredentialStoreSQL(db, p.Secrets)
		}
		if p.Parameters == nil {
			p.Parameters = storage.NewParameterStoreSQL(db, p.Secrets)
		}

		return nil
	}

	// Load the config file and replace any referenced secrets
	ctx, err := p.Config.Load(ctx, func(innerCtx context.Context, secret string) (string, error) {
		value, err := p.Secrets.Resolve(innerCtx, "secret", secret)
		if err != nil {
			if strings.Contains(err.Error(), "invalid value source: secret") {
				return "", errors.New("No secret store account is configured")
			}
			return "", err
		}
		return value, nil
	})
	if err != nil {
		return ctx, err
	}

	if err = init(ctx); err != nil {
		return ctx, err
	}

	if p.Sanitizer == nil {
		p.Sanitizer = storage.NewSanitizer(p.Parameters, p.Secrets)
	}
	if p.CNAB == nil {
		p.CNAB = cnabprovider.NewRuntime(p.Config, p.Installations, p.Credentials, p.Parameters, p.Secrets, p.Sanitizer)
	}

	return ctx, nil
}

// Close releases resources used by Porter before terminating the application.
func (p *Porter) Close() error {
	// Shutdown our plugins
	var bigErr *multierror.Error

	err := p.Secrets.Close()
	if err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	if p.Storage != nil {
		err = p.Storage.Close()
		if err != nil {
			bigErr = multierror.Append(bigErr, err)
		}
	}
	for _, do := range p.onClose {
		if err = do(); err != nil {
			bigErr = multierror.Append(bigErr, err)
		}
	}

	err = p.Config.Close()
	if err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	err = p.Signer.Close()
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
