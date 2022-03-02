package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	portercontext "get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/pkg/errors"
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

	// EnvCORRELATION_ID is the name of the environment variable containing the
	// id to correlate logs with a workflow.
	EnvCorrelationID = "PORTER_CORRELATION_ID"

	// CustomPorterKey is the key in the bundle.json custom section that contains the Porter stamp
	// It holds all the metadata that Porter includes that is specific to Porter about the bundle.
	CustomPorterKey = "sh.porter"

	// BundleOutputsDir is the directory where outputs are expected to be placed
	// during the execution of a bundle action.
	BundleOutputsDir = "/cnab/app/outputs"

	// ClaimFilepath is the filepath to the claim.json inside of an invocation image
	ClaimFilepath = "/cnab/claim.json"
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
}

// New Config initializes a default porter configuration.
func New() *Config {
	return &Config{
		Context:    portercontext.New(),
		Data:       DefaultDataStore(),
		DataLoader: LoadFromEnvironment(),
	}
}

// loadData from the datastore defined in PORTER_HOME, and render the
// config file using the specified template data.
func (c *Config) loadData(ctx context.Context, templateData map[string]interface{}) error {
	if c.DataLoader == nil {
		c.DataLoader = LoadFromEnvironment()
	}

	if err := c.DataLoader(ctx, c, templateData); err != nil {
		return err
	}

	if c.IsFeatureEnabled(experimental.FlagStructuredLogs) {
		// Now that we have completely loaded our config, configure our final logging/tracing
		c.Context.ConfigureLogging(portercontext.LogConfiguration{
			StructuredLogs:       true,
			LogToFile:            c.Data.Logs.Enabled,
			LogDirectory:         filepath.Join(c.porterHome, "logs"),
			LogLevel:             c.Data.Logs.Level.Level(),
			LogCorrelationID:     c.Getenv(EnvCorrelationID),
			TelemetryEnabled:     c.Data.Telemetry.Enabled,
			TelemetryEndpoint:    c.Data.Telemetry.Endpoint,
			TelemetryProtocol:    c.Data.Telemetry.Protocol,
			TelemetryInsecure:    c.Data.Telemetry.Insecure,
			TelemetryCertificate: c.Data.Telemetry.Certificate,
			TelemetryCompression: c.Data.Telemetry.Compression,
			TelemetryTimeout:     c.Data.Telemetry.Timeout,
			TelemetryHeaders:     c.Data.Telemetry.Headers,
		})
	}

	return nil
}

func (c *Config) GetStorage(name string) (StoragePlugin, error) {
	if c != nil {
		for _, is := range c.Data.StoragePlugins {
			if is.Name == name {
				return is, nil
			}
		}
	}

	return StoragePlugin{}, errors.Errorf("store '%s' not defined", name)
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
			return "", errors.Wrap(err, "could not get user home directory")
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

func (c *Config) GetPorterPath() (string, error) {
	if c.porterPath != "" {
		return c.porterPath, nil
	}

	porterPath, err := getExecutable()
	if err != nil {
		return "", errors.Wrap(err, "could not get path to the executing porter binary")
	}

	// We try to resolve back to the original location
	hardPath, err := evalSymlinks(porterPath)
	if err != nil { // if we have trouble resolving symlinks, skip trying to help people who used symlinks
		fmt.Fprintln(c.Err, errors.Wrapf(err, "WARNING could not resolve %s for symbolic links\n", porterPath))
	} else if hardPath != porterPath {
		if c.Debug {
			fmt.Fprintf(c.Err, "Resolved porter binary from %s to %s\n", porterPath, hardPath)
		}
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
// Example: Config.SetExperimentalFlags(experimental.FlagBuildDrivers | ...)
func (c *Config) SetExperimentalFlags(flags experimental.FeatureFlags) {
	c.experimental = &flags
}

// GetBuildDriver determines the correct build driver to use, taking
// into account experimental flags.
// Use this instead of Config.Data.BuildDriver directly.
func (c *Config) GetBuildDriver() string {
	if c.IsFeatureEnabled(experimental.FlagBuildDrivers) {
		return c.Data.BuildDriver
	}
	return BuildDriverDocker
}

// Load loads the configuration file, rendering any templating used in the config file
// such as ${secret.NAME} or ${env.NAME}.
// Pass nil for resolveSecret to skip resolving secrets.
func (c *Config) Load(ctx context.Context, resolveSecret func(secretKey string) (string, error)) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	if err := c.loadFirstPass(ctx); err != nil {
		return err
	}

	if err := c.loadFinalPass(ctx, resolveSecret); err != nil {
		return err
	}

	// Record some global configuration values that are relevant to most commands
	log.SetAttributes(
		attribute.String("porter.config.namespace", c.Data.Namespace),
		attribute.String("porter.config.experimental", strings.Join(c.Data.ExperimentalFlags, ",")),
	)

	return nil
}

// our first pass only loads the config file while replacing
// environment variables. Once we have that we can use the
// config to connect to a secret store and do a second pass
// over the config.
func (c *Config) loadFirstPass(ctx context.Context) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	templateData := map[string]interface{}{
		"env": c.EnvironMap(),
	}
	return c.loadData(ctx, templateData)
}

func (c *Config) loadFinalPass(ctx context.Context, resolveSecret func(secretKey string) (string, error)) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	// Don't do extra work if there aren't any secrets
	if len(c.templateVariables) == 0 || resolveSecret == nil {
		return nil
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

			_, childLog := log.StartSpanWithName("resolveSecret", attribute.String("porter.config.secret.key", secretKey))
			defer childLog.EndSpan()
			secretValue, err := resolveSecret(secretKey)
			if err != nil {
				return childLog.Error(errors.Wrapf(err, "could not render config file because ${secret.%s} could not be resolved", secretKey))
			}

			secrets[secretKey] = secretValue
			return nil
		}(variable)
		if err != nil {
			return err
		}
	}

	templateData := map[string]interface{}{
		"env":    c.EnvironMap(),
		"secret": secrets,
	}

	// reload configuration with secrets loaded
	return c.loadData(ctx, templateData)
}
