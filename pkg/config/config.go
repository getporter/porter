package config

import (
	"fmt"
	"os"
	"path/filepath"

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
)

type Config struct {
	*context.Context
	Manifest *Manifest
}

// New Config initializes a default porter configuration.
func New() *Config {
	return &Config{
		Context: context.New(),
	}
}

// GetHomeDir determines the path to the porter home directory.
func (c *Config) GetHomeDir() (string, error) {
	home, ok := os.LookupEnv(EnvHOME)
	if ok {
		return home, nil
	}

	porterPath, err := os.Executable()
	if err != nil {
		return "", errors.Wrap(err, "could not get path to the executing porter binary")
	}

	porterDir := filepath.Dir(porterPath)

	return porterDir, nil
}

// GetTemplatesDir determines the path to the templates directory.
func (c *Config) GetTemplatesDir() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "templates"), nil
}

// GetPorterConfigTemplate reads templates/porter.yaml from the porter home directory.
func (c *Config) GetPorterConfigTemplate() ([]byte, error) {
	tmplDir, err := c.GetTemplatesDir()
	if err != nil {
		return nil, err
	}

	tmplPath := filepath.Join(tmplDir, Name)
	return c.FileSystem.ReadFile(tmplPath)
}

// GetRunScriptTemplate reads templates/run from the porter home directory.
func (c *Config) GetRunScriptTemplate() ([]byte, error) {
	tmplDir, err := c.GetTemplatesDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(tmplDir, filepath.Base(RunScript))
	b, err := c.FileSystem.ReadFile(path)
	return b, errors.Wrapf(err, "could not read script template at %s", path)
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
	// Check for a local bundle next to the current manifest
	if c.Manifest != nil {
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
