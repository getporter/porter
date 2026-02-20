package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/schema"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// Name is the file name of the porter configuration file.
	Name = "porter.yaml"

	// EnvHOME is the name of the environment variable containing the porter home directory path.
	EnvHOME = "PORTER_HOME"

	// EnvBundleName is the name of the environment variable containing the name of the bundle.
	EnvBundleName = "CNAB_BUNDLE_NAME"

	// EnvInstallationName is the name of the environment variable containing the name of the installation.
	EnvInstallationName = "CNAB_INSTALLATION_NAME"

	// EnvACTION is the requested action to be executed
	EnvACTION = "CNAB_ACTION"

	// EnvDEBUG is a custom porter parameter that signals that --debug flag has been passed through from the client to the runtime.
	EnvDEBUG = "PORTER_DEBUG"

	// CustomPorterKey is the key in the bundle.json custom section that contains the Porter stamp
	// It holds all the metadata that Porter includes that is specific to Porter about the bundle.
	CustomPorterKey = "sh.porter"

	// BundleOutputsDir is the directory where outputs are expected to be placed
	// during the execution of a bundle action.
	BundleOutputsDir = "/cnab/app/outputs"

	// ClaimFilepath is the filepath to the claim.json inside of an bundle image
	ClaimFilepath = "/cnab/claim.json"

	// EnvPorterInstallationNamespace is the name of the environment variable which is injected into the
	// bundle image, containing the namespace of the installation.
	EnvPorterInstallationNamespace = "PORTER_INSTALLATION_NAMESPACE"

	// EnvPorterInstallationName is the name of the environment variable which is injected into the
	// bundle image, containing the name of the installation.
	EnvPorterInstallationName = "PORTER_INSTALLATION_NAME"

	// EnvPorterInstallationID is the name of the environment variable which is injected into the
	// bundle image, containing the unique ID of the installation.
	EnvPorterInstallationID = "PORTER_INSTALLATION_ID"

	// DefaultVerbosity is the default value for the --verbosity flag.
	DefaultVerbosity = "info"
)

// These are functions that afero doesn't support, so this lets us stub them out for tests to set the
// location of the current executable porter binary and resolve PORTER_HOME.
var getExecutable = os.Executable
var evalSymlinks = filepath.EvalSymlinks

// DataStoreLoaderFunc defines the Config.DataLoader function signature
// used to load data into Config.DataStore.
type DataStoreLoaderFunc func(context.Context, *Config, map[string]interface{}) error

type Config struct {
	*portercontext.Context
	Data       Data
	DataLoader DataStoreLoaderFunc

	// ConfigFilePath is the path to the loaded configuration file
	ConfigFilePath string

	// Cache the resolved Porter home directory
	porterHome string

	// Cache the resolved Porter binary path
	porterPath string

	// parsed feature flags
	experimental *experimental.FeatureFlags

	// list of variables used in the config file
	// for example: secret.NAME, or env.NAME
	templateVariables []string

	// the populated viper instance that loaded the current configuration
	viper *viper.Viper
}

// New Config initializes a default porter configuration.
func New() *Config {
	return NewFor(portercontext.New())
}

// NewFor initializes a porter configuration, using an existing porter context.
func NewFor(pCtx *portercontext.Context) *Config {
	return &Config{
		Context:    pCtx,
		Data:       DefaultDataStore(),
		DataLoader: LoadFromEnvironment(),
	}
}

func (c *Config) NewLogConfiguration() portercontext.LogConfiguration {
	return portercontext.LogConfiguration{
		Verbosity:               c.GetVerbosity().Level(),
		StructuredLogs:          c.Data.Logs.Structured,
		LogToFile:               c.Data.Logs.LogToFile,
		LogDirectory:            filepath.Join(c.porterHome, "logs"),
		LogLevel:                c.Data.Logs.Level.Level(),
		TelemetryEnabled:        c.Data.Telemetry.Enabled,
		TelemetryEndpoint:       c.Data.Telemetry.Endpoint,
		TelemetryProtocol:       c.Data.Telemetry.Protocol,
		TelemetryInsecure:       c.Data.Telemetry.Insecure,
		TelemetryCertificate:    c.Data.Telemetry.Certificate,
		TelemetryCompression:    c.Data.Telemetry.Compression,
		TelemetryTimeout:        c.Data.Telemetry.Timeout,
		TelemetryHeaders:        c.Data.Telemetry.Headers,
		TelemetryServiceName:    "porter",
		TelemetryDirectory:      filepath.Join(c.porterHome, "traces"),
		TelemetryRedirectToFile: c.Data.Telemetry.RedirectToFile,
		TelemetryStartTimeout:   c.Data.Telemetry.GetStartTimeout(),
	}
}

// loadData from the datastore defined in PORTER_HOME, and render the
// config file using the specified template data.
func (c *Config) loadData(ctx context.Context, templateData map[string]interface{}) (context.Context, error) {
	if c.DataLoader == nil {
		c.DataLoader = LoadFromEnvironment()
	}

	if err := c.DataLoader(ctx, c, templateData); err != nil {
		return ctx, err
	}

	// Now that we have completely loaded our config, configure our final logging/tracing
	ctx = c.Context.ConfigureLogging(ctx, c.NewLogConfiguration())
	return ctx, nil
}

func (c *Config) GetSchemaCheckStrategy(ctx context.Context) schema.CheckStrategy {
	switch c.Data.SchemaCheck {
	case string(schema.CheckStrategyMinor):
		return schema.CheckStrategyMinor
	case string(schema.CheckStrategyMajor):
		return schema.CheckStrategyMajor
	case string(schema.CheckStrategyNone):
		return schema.CheckStrategyNone
	case string(schema.CheckStrategyExact), "":
		return schema.CheckStrategyExact
	default:
		log := tracing.LoggerFromContext(ctx)
		log.Warnf("invalid schema-check value specified %q, defaulting to exact", c.Data.SchemaCheck)
		return schema.CheckStrategyExact
	}
}

func (c *Config) GetStorage(name string) (StoragePlugin, error) {
	if c != nil {
		for _, is := range c.Data.StoragePlugins {
			if is.Name == name {
				return is, nil
			}
		}
	}

	return StoragePlugin{}, fmt.Errorf("store '%s' not defined", name)
}

func (c *Config) GetSecretsPlugin(name string) (SecretsPlugin, error) {
	if c != nil {
		for _, cs := range c.Data.SecretsPlugin {
			if cs.Name == name {
				return cs, nil
			}
		}
	}

	return SecretsPlugin{}, errors.New("secrets %q not defined")
}

func (c *Config) GetSigningPlugin(name string) (SigningPlugin, error) {
	if c != nil {
		for _, cs := range c.Data.SigningPlugin {
			if cs.Name == name {
				return cs, nil
			}
		}
	}

	return SigningPlugin{}, errors.New("signing %q not defined")
}

// GetHomeDir determines the absolute path to the porter home directory.
// Hierarchy of checks:
// - PORTER_HOME
// - HOME/.porter or USERPROFILE/.porter
func (c *Config) GetHomeDir() (string, error) {
	if c.porterHome != "" {
		return c.porterHome, nil
	}

	home := c.Getenv(EnvHOME)
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get user home directory: %w", err)
		}
		home = filepath.Join(userHome, ".porter")
	}

	// As a relative path may be supplied via EnvHOME,
	// we want to return the absolute path for programmatic usage elsewhere,
	// for instance, in setting up volume mounts for outputs
	c.SetHomeDir(c.FileSystem.Abs(home))

	return c.porterHome, nil
}

// SetHomeDir is a test function that allows tests to use an alternate
// Porter home directory.
func (c *Config) SetHomeDir(home string) {
	c.porterHome = home

	// Set this as an environment variable so that when we spawn new processes
	// such as a mixin or plugin, that they can find PORTER_HOME too
	c.Setenv(EnvHOME, home)
}

// SetPorterPath is a test function that allows tests to use an alternate
// Porter binary location.
func (c *Config) SetPorterPath(path string) {
	c.porterPath = path
}

func (c *Config) GetPorterPath(ctx context.Context) (string, error) {
	if c.porterPath != "" {
		return c.porterPath, nil
	}

	log := tracing.LoggerFromContext(ctx)
	porterPath, err := getExecutable()
	if err != nil {
		return "", log.Error(fmt.Errorf("could not get path to the executing porter binary: %w", err))
	}

	// We try to resolve back to the original location
	hardPath, err := evalSymlinks(porterPath)
	if err != nil { // if we have trouble resolving symlinks, skip trying to help people who used symlinks
		return "", log.Error(fmt.Errorf("WARNING could not resolve %s for symbolic links: %w", porterPath, err))
	}
	if hardPath != porterPath {
		log.Debugf("Resolved porter binary from %s to %s", porterPath, hardPath)
		porterPath = hardPath
	}

	c.porterPath = porterPath
	return porterPath, nil
}

// GetBundlesCache locates the bundle cache from the porter home directory.
func (c *Config) GetBundlesCache() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "bundles"), nil
}

func (c *Config) GetPluginsDir() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "plugins"), nil
}

func (c *Config) GetPluginPath(plugin string) (string, error) {
	pluginsDir, err := c.GetPluginsDir()
	if err != nil {
		return "", err
	}

	executablePath := filepath.Join(pluginsDir, plugin, plugin)
	return executablePath, nil
}

// GetBundleArchiveLogs locates the output for Bundle Archive Operations.
func (c *Config) GetBundleArchiveLogs() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "archives"), nil
}

// GetFeatureFlags indicates which experimental feature flags are enabled
func (c *Config) GetFeatureFlags() experimental.FeatureFlags {
	if c.experimental == nil {
		flags := experimental.ParseFlags(c.Data.ExperimentalFlags)
		c.experimental = &flags
	}
	return *c.experimental
}

// IsFeatureEnabled returns true if the specified experimental flag is enabled.
func (c *Config) IsFeatureEnabled(flag experimental.FeatureFlags) bool {
	return c.GetFeatureFlags()&flag == flag
}

// SetExperimentalFlags programmatically, overriding Config.Data.ExperimentalFlags.
// Example: Config.SetExperimentalFlags(experimental.FlagStructuredLogs | ...)
func (c *Config) SetExperimentalFlags(flags experimental.FeatureFlags) {
	c.experimental = &flags
}

// GetBuildDriver determines the correct build driver to use, taking
// into account experimental flags.
// Use this instead of Config.Data.BuildDriver directly.
func (c *Config) GetBuildDriver() string {
	return BuildDriverBuildkit
}

// GetVerbosity converts the user-specified verbosity flag into a LogLevel enum.
func (c *Config) GetVerbosity() LogLevel {
	return ParseLogLevel(c.Data.Verbosity)
}

// Load loads the configuration file, rendering any templating used in the config file
// such as ${secret.NAME} or ${env.NAME}.
// Pass nil for resolveSecret to skip resolving secrets.
func (c *Config) Load(ctx context.Context, resolveSecret func(ctx context.Context, secretKey string) (string, error)) (context.Context, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	ctx, err := c.loadFirstPass(ctx)
	if err != nil {
		return ctx, err
	}

	ctx, err = c.loadFinalPass(ctx, resolveSecret)
	if err != nil {
		return ctx, err
	}

	// Record some global configuration values that are relevant to most commands
	log.SetAttributes(
		attribute.String("porter.config.namespace", c.Data.Namespace),
		attribute.String("porter.config.experimental", strings.Join(c.Data.ExperimentalFlags, ",")),
	)

	return ctx, nil
}

// our first pass only loads the config file while replacing
// environment variables. Once we have that we can use the
// config to connect to a secret store and do a second pass
// over the config.
func (c *Config) loadFirstPass(ctx context.Context) (context.Context, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	templateData := map[string]interface{}{
		"env": c.EnvironMap(),
	}
	return c.loadData(ctx, templateData)
}

func (c *Config) loadFinalPass(ctx context.Context, resolveSecret func(ctx context.Context, secretKey string) (string, error)) (context.Context, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	// Don't do extra work if there aren't any secrets
	if len(c.templateVariables) == 0 || resolveSecret == nil {
		return ctx, nil
	}

	secrets := make(map[string]string, len(c.templateVariables))
	for _, variable := range c.templateVariables {
		err := func(variable string) error {
			// Check if it's a secret variable, e.g. ${secret.NAME}
			secretPrefix := "secret."
			i := strings.Index(variable, secretPrefix)
			if i == -1 {
				return nil
			}

			secretKey := variable[len(secretPrefix):]

			ctx, childLog := log.StartSpanWithName("resolveSecret", attribute.String("porter.config.secret.key", secretKey))
			defer childLog.EndSpan()
			secretValue, err := resolveSecret(ctx, secretKey)
			if err != nil {
				return childLog.Error(fmt.Errorf("could not render config file because ${secret.%s} could not be resolved: %w", secretKey, err))
			}

			secrets[secretKey] = secretValue
			return nil
		}(variable)
		if err != nil {
			return ctx, err
		}
	}

	templateData := map[string]interface{}{
		"env":    c.EnvironMap(),
		"secret": secrets,
	}

	// reload configuration with secrets loaded
	return c.loadData(ctx, templateData)
}

// ExportRemoteConfigAsEnvironmentVariables represents the current configuration
// as environment variables suitable for a remote Porter actor, such as a mixin
// or plugin. Only a subset of values are exported, such as tracing and logging,
// and not plugin configuration (since it's not relevant when running a plugin
// and may contain sensitive data). For example, if Config.Data.Logs is set to warn, it
// would return PORTER_LOGS_LEVEL=warn in the resulting set of environment variables.
// This is used to pass config from porter to a mixin or plugin.
func (c *Config) ExportRemoteConfigAsEnvironmentVariables() []string {
	if c.viper == nil {
		return nil
	}

	// the set of config that is relevant to remote actors
	keepPrefixes := []string{"verbosity", "logs", "telemetry"}

	var env []string
	for _, key := range c.viper.AllKeys() {
		for _, prefix := range keepPrefixes {
			if strings.HasPrefix(key, prefix) {
				val := c.viper.Get(key)
				if reflect.ValueOf(val).IsZero() {
					continue
				}
				envVarSuffix := strings.ToUpper(key)
				envVarSuffix = strings.NewReplacer(".", "_", "-", "_").
					Replace(envVarSuffix)
				envVar := fmt.Sprintf("PORTER_%s", envVarSuffix)
				env = append(env, fmt.Sprintf("%s=%v", envVar, val))
			}
		}
	}

	return env
}
