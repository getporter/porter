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

	// EnvHOME is the name of the environment variable containing the porter home directory path.
	EnvHOME = "PORTER_HOME"

	// EnvBundleName is the name of the environment variable containing the name of the bundle.
	EnvBundleName = "CNAB_ACTION"

	// EnvClaimName is the name of the environment variable containing the name of the claim.
	EnvClaimName = "CNAB_INSTALLATION_NAME"

	// EnvACTION is the request
	EnvACTION = "CNAB_ACTION"

	// EnvDEBUG is a custom porter parameter that signals that --debug flag has been passed through from the client to the runtime.
	EnvDEBUG = "PORTER_DEBUG"

	CustomBundleKey = "sh.porter"

	// BundleOutputsDir is the directory where outputs are expected to be placed
	// during the execution of a bundle action.
	BundleOutputsDir = "/cnab/app/outputs"
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

// GetHomeDir determines the absolute path to the porter home directory.
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

	// As a relative path may be supplied via EnvHOME,
	// we want to return the absolute path for programmatic usage elsewhere,
	// for instance, in setting up volume mounts for outputs
	absoluteHome, err := filepath.Abs(home)
	if err != nil {
		return "", errors.Wrap(err, "could not get the absolute path for the porter home directory")
	}

	c.porterHome = absoluteHome
	return c.porterHome, nil
}

func (c *Config) SetHomeDir(home string) {
	c.porterHome = home
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

// GetBundlesDir locates the bundle cache from the porter home directory.
func (c *Config) GetBundlesCache() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "bundles"), nil
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

func (c *Config) GetCredentialsDir() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "credentials"), nil
}

func (c *Config) GetCredentialPath(name string) (string, error) {
	credDir, err := c.GetCredentialsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(credDir, fmt.Sprintf("%s.yaml", name)), nil
}

func (c *Config) GetOutputsDir() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "outputs"), nil
}
