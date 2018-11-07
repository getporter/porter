package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	// Name is the file name of the porter configuration file.
	Name    = "porter.yaml"
	EnvHOME = "PORTER_HOME"
)

// GetHomeDir determines the path to the porter home directory.
func GetHomeDir() (string, error) {
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
func GetTemplatesDir() (string, error) {
	home, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "templates"), nil
}

// GetPorterConfigTemplate reads templates/porter.yaml from the porter home directory.
func GetPorterConfigTemplate() ([]byte, error) {
	tmplDir, err := GetTemplatesDir()
	if err != nil {
		return nil, err
	}

	tmplPath := filepath.Join(tmplDir, Name)
	return ioutil.ReadFile(tmplPath)
}
