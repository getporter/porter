package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// Name is the file name of the porter configuration file.
	Name    = "porter.yaml"
	EnvHOME = "PORTER_HOME"
)

type Config struct {
	FileSystem *afero.Afero
	Out        io.Writer
}

// New Config initializes a default porter configuration.
func New() Config {
	return Config{
		FileSystem: &afero.Afero{Fs: afero.NewOsFs()},
		Out:        os.Stdout,
	}
}

// NewTestConfig initializes a configuration suitable for testing, with the output buffered, and an in-memory file system.
func NewTestConfig() (Config, *bytes.Buffer) {
	output := &bytes.Buffer{}
	c := Config{
		FileSystem: &afero.Afero{Fs: afero.NewMemMapFs()},
		Out:        output,
	}

	return c, output
}

// GetHomeDir determines the path to the porter home directory.
func (c Config) GetHomeDir() (string, error) {
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
func (c Config) GetTemplatesDir() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "templates"), nil
}

// GetPorterConfigTemplate reads templates/porter.yaml from the porter home directory.
func (c Config) GetPorterConfigTemplate() ([]byte, error) {
	tmplDir, err := c.GetTemplatesDir()
	if err != nil {
		return nil, err
	}

	tmplPath := filepath.Join(tmplDir, Name)
	return c.FileSystem.ReadFile(tmplPath)
}
