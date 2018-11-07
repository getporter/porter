package porter

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/config"
	"github.com/stretchr/testify/require"
)

// NewTestPorter initializes a porter test client, with the output buffered, and an in-memory file system.
func NewTestPorter(t *testing.T) (*Porter, *bytes.Buffer) {
	testConfig, output := config.NewTestConfig()
	p := &Porter{
		Config: testConfig,
	}

	return p, output
}

// InitializePorterHome initializes the test filesystem with the supporting files in the PORTER_HOME directory.
func InitializePorterHome(t *testing.T, p *Porter) {
	// Set up the test porter home directory
	os.Setenv(config.EnvHOME, "/root/.porter")
	tmplDir, err := p.Config.GetTemplatesDir()
	require.NoError(t, err)
	err = p.FileSystem.Mkdir(tmplDir, os.ModePerm)
	require.NoError(t, err)

	// Setup the pwd
	pwd, err := filepath.Abs(".")
	require.NoError(t, err)
	err = p.FileSystem.MkdirAll(pwd, os.ModePerm)
	require.NoError(t, err)

	// Copy templates
	tmpl, err := ioutil.ReadFile("testdata/porter.yaml")
	require.NoError(t, err)

	err = p.FileSystem.WriteFile(filepath.Join(tmplDir, config.Name), tmpl, os.ModePerm)
	require.NoError(t, err)
}
