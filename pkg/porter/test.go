package porter

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

	// Copy templates
	srcDir := "../../templates"
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		tmpl, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(tmplDir, strings.TrimPrefix(path, srcDir))
		return p.FileSystem.WriteFile(destPath, tmpl, os.ModePerm)
	})
	require.NoError(t, err)
}
