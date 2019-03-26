package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

const (
	// Name is the file name of the porter configuration file.
	Name = "porter.yaml"

	// RunScript is the path to the CNAB run script.
	RunScript = "cnab/app/run"

	// EnvHOME is the name of the environment variable containing the porter home directory path.
	EnvHOME = "PORTER_HOME"

	// EnvACTION is the request
	EnvACTION = "CNAB_ACTION"

	// EnvDEBUG is a custom porter parameter that signals that --debug flag has been passed through from the client to the runtime.
	EnvDEBUG = "PORTER_DEBUG"
)

// These are functions that afero doesn't support, so this lets us stub them out for tests to set the
// location of the current executable porter binary and resolve PORTER_HOME.
var getExecutable = os.Executable
var evalSymlinks = filepath.EvalSymlinks

type Config struct {
	*context.Context
	Manifest *Manifest

	porterHome string
}

// New Config initializes a default porter configuration.
func New() *Config {
	return &Config{
		Context: context.New(),
	}
}

// GetHomeDir determines the path to the porter home directory.
func (c *Config) GetHomeDir() (string, error) {
	if c.porterHome != "" {
		return c.porterHome, nil
	}

	home := os.Getenv(EnvHOME)

	if home == "" {
		porterPath, err := getExecutable()
		if err != nil {
			return "", errors.Wrap(err, "could not get path to the executing porter binary")
		}

		// This is for the scenario when someone symlinks the ~/.porter/porter binary to /usr/local/porter
		// We try to resolve back to the original location so that we can find the mixins, etc next to it.
		hardPath, err := evalSymlinks(porterPath)
		if err != nil { // if we have trouble resolving symlinks, skip trying to help people who used symlinks
			fmt.Fprintln(c.Err, errors.Wrapf(err, "WARNING could not resolve %s for symbolic links\n", porterPath))
		} else if hardPath != porterPath {
			if c.Debug {
				fmt.Fprintf(c.Err, "Resolved porter binary from %s to %s\n", porterPath, hardPath)
			}
			porterPath = hardPath
		}
		home = filepath.Dir(porterPath)
	}

	c.porterHome = home
	return c.porterHome, nil
}

func (c *Config) GetPorterPath() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}

	executablePath := filepath.Join(home, "porter")
	return executablePath, nil
}

func (c *Config) GetPorterRuntimePath() (string, error) {
	path, err := c.GetPorterPath()
	if err != nil {
		return "", nil
	}

	return path + "-runtime", nil
}

// GetBundleManifest gets the path to another bundle's manifest.
func (c *Config) GetBundleManifestPath(bundle string) (string, error) {
	bundlesDir, err := c.GetBundleDir(bundle)
	if err != nil {
		return "", err
	}

	return filepath.Join(bundlesDir, Name), nil
}

// GetBundlesDir locates the bundle cache from the porter home directory.
func (c *Config) GetBundlesCache() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "bundles"), nil
}

// GetBundleDir locates a bundle
// Lookup order:
// - ./bundles/
// - PORTER_HOME/bundles/
func (c *Config) GetBundleDir(bundle string) (string, error) {
	urlPath := strings.HasPrefix(c.Manifest.path, "http")

	// Check for a local bundle next to the current manifest
	if c.Manifest != nil || urlPath == false {
		localDir := c.Manifest.GetManifestDir()
		localBundleDir := filepath.Join(localDir, "bundles", bundle)
		dirExists, err := c.FileSystem.DirExists(localBundleDir)

		if err != nil {
			return "", errors.Wrapf(err, "could not check if directory %s exists", localBundleDir)
		}

		if dirExists {
			return localBundleDir, nil
		}
	}

	// Fall back to looking in the cache under PORTER_HOME
	cacheDir, err := c.GetBundlesCache()
	if err != nil {
		return "", err
	}

	bundleDir := filepath.Join(cacheDir, bundle)

	dirExists, err := c.FileSystem.DirExists(bundleDir)
	if err != nil {
		return "", errors.Wrapf(err, "bundle %s not accessible at %s", bundle, bundleDir)
	}
	if !dirExists {
		h, _ := c.GetHomeDir()
		return "", errors.Errorf("bundle %s not available locally or in %s", bundle, h)
	}

	return bundleDir, nil
}

func (c *Config) GetMixinsDir() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "mixins"), nil
}

func (c *Config) GetMixinDir(mixin string) (string, error) {
	mixinsDir, err := c.GetMixinsDir()
	if err != nil {
		return "", err
	}

	mixinDir := filepath.Join(mixinsDir, mixin)

	dirExists, err := c.FileSystem.DirExists(mixinDir)
	if err != nil {
		return "", errors.Wrapf(err, "mixin %s not accessible at %s", mixin, mixinDir)
	}
	if !dirExists {
		return "", fmt.Errorf("mixin %s not installed in PORTER_HOME", mixin)
	}

	return mixinDir, nil
}

func (c *Config) GetMixinPath(mixin string) (string, error) {
	mixinDir, err := c.GetMixinDir(mixin)
	if err != nil {
		return "", err
	}

	executablePath := filepath.Join(mixinDir, mixin)
	return executablePath, nil
}

func (c *Config) GetMixinRuntimePath(mixin string) (string, error) {
	path, err := c.GetMixinPath(mixin)
	if err != nil {
		return "", nil
	}

	return path + "-runtime", nil
}
