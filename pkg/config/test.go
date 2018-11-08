package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func SetupPorterHome(t *testing.T, c *Config) map[string][]byte {
	templates := make(map[string][]byte, 2)

	// Set up a test porter home
	testHome := "/root/.porter"
	err := c.FileSystem.MkdirAll(testHome, os.ModePerm)
	require.NoError(t, err)
	os.Setenv(EnvHOME, testHome)

	// Setup a templates directory
	templatesDir, err := c.GetTemplatesDir()
	require.NoError(t, err)
	err = c.FileSystem.Mkdir(templatesDir, os.ModePerm)
	require.NoError(t, err)

	// Add a template porter.yaml
	templates["porter.yaml"] = CopyTemplate(t, c, "../../templates/run", Name)

	// Add a template run script
	templates["run"] = CopyTemplate(t, c, "../../templates/porter.yaml", "run")

	return templates
}

func CopyTemplate(t *testing.T, c *Config, src, dest string) []byte {
	templatesDir, err := c.GetTemplatesDir()
	require.NoError(t, err)

	templDest := filepath.Join(templatesDir, dest)
	return CopyFile(t, c, src, templDest)
}

func CopyFile(t *testing.T, c *Config, src, dest string) []byte {
	data, err := ioutil.ReadFile(src)
	require.NoError(t, err)

	err = c.FileSystem.WriteFile(dest, data, os.ModePerm)
	require.NoError(t, err)

	return data
}
