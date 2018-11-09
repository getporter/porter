package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deislabs/porter/pkg/context"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	*Config
	TestContext *context.TestContext

	Templates map[string][]byte
}

// NewTestConfig initializes a configuration suitable for testing, with the output buffered, and an in-memory file system.
func NewTestConfig(t *testing.T) *TestConfig {
	cxt := context.NewTestContext(t)
	c := &TestConfig{
		Config: &Config{
			Context: cxt.Context,
		},
		TestContext: cxt,
		Templates:   make(map[string][]byte),
	}
	return c
}

// InitializePorterHome initializes the test filesystem with the supporting files in the PORTER_HOME directory.
func (c *TestConfig) SetupPorterHome() {
	// Set up the test porter home directory
	os.Setenv(EnvHOME, "/root/.porter")

	// Copy templates
	srcDir := "../../templates/"
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		dest := strings.TrimPrefix(path, srcDir)
		c.AddTemplate(path, dest)

		return nil
	})
	require.NoError(c.TestContext.T, err)
}

func (c *TestConfig) AddTemplate(src, dest string) {
	templatesDir, err := c.GetTemplatesDir()
	require.NoError(c.TestContext.T, err)

	templDest := filepath.Join(templatesDir, dest)
	tmpl := c.TestContext.AddFile(src, templDest)
	c.Templates[dest] = tmpl
}
