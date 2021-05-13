package config

import (
	"fmt"
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/experimental"
	"github.com/pkg/errors"
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

	// ClaimFilepath is the filepath to the claim.json inside of an invocation image
	ClaimFilepath = "/cnab/claim.json"
)

// These are functions that afero doesn't support, so this lets us stub them out for tests to set the
// location of the current executable porter binary and resolve PORTER_HOME.
var getExecutable = os.Executable
var evalSymlinks = filepath.EvalSymlinks

type DataStoreLoaderFunc func(*Config) error

type Config struct {
	*context.Context
	Data       Data
	DataLoader DataStoreLoaderFunc

	// Cache the resolved Porter home directory
	porterHome string

	// Cache the resolved Porter binary path
	porterPath string

	// parsed feature flags
	experimental *experimental.FeatureFlags
}

// New Config initializes a default porter configuration.
func New() *Config {
	return &Config{
		Context:    context.New(),
		Data:       DefaultDataStore(),
		DataLoader: LoadFromEnvironment(),
	}
}

// LoadData from the datastore in PORTER_HOME.
// This defaults to reading the configuration file and environment variables.
func (c *Config) LoadData() error {
	if c.DataLoader == nil {
		c.DataLoader = LoadFromEnvironment()
	}

	return c.DataLoader(c)
}

func (c *Config) GetStorage(name string) (CrudStore, error) {
	if c != nil {
		for _, is := range c.Data.CrudStores {
			if is.Name == name {
				return is, nil
			}
		}
	}

	return CrudStore{}, errors.New("store %q not defined")
}

func (c *Config) GetSecretSource(name string) (SecretSource, error) {
	if c != nil {
		for _, cs := range c.Data.SecretSources {
			if cs.Name == name {
				return cs, nil
			}
		}
	}

	return SecretSource{}, errors.New("secrets %q not defined")
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

// GetBundlesDir locates the bundle cache from the porter home directory.
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

// GetArchiveLogs locates the output for Bundle Archive Operations.
func (c *Config) GetBundleArchiveLogs() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "archives"), nil
}

// FeatureFlags indicates which experimental feature flags are enabled
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
