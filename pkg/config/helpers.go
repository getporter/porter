package config

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/context"
)

type TestConfig struct {
	*Config
	TestContext *context.TestContext
}

// NewTestConfig initializes a configuration suitable for testing, with the output buffered, and an in-memory file system.
func NewTestConfig(t *testing.T) *TestConfig {
	tc := context.NewTestContext(t)
	cfg := New()
	cfg.Context = tc.Context
	return &TestConfig{
		Config:      cfg,
		TestContext: tc,
	}
}

// InitializePorterHome initializes the test filesystem with the supporting files in the PORTER_HOME directory.
func (c *TestConfig) SetupPorterHome() {
	// Set up the test porter home directory
	home := "/root/.porter"
	c.SetHomeDir(home)

	// Fake out the porter home directory
	c.FileSystem.Create(filepath.Join(home, "porter"))
	c.FileSystem.Create(filepath.Join(home, "porter-runtime"))

	mixinsDir := filepath.Join(home, "mixins")
	c.FileSystem.Create(filepath.Join(mixinsDir, "exec/exec"))
	c.FileSystem.Create(filepath.Join(mixinsDir, "exec/exec-runtime"))
	c.FileSystem.Create(filepath.Join(mixinsDir, "helm/helm"))
	c.FileSystem.Create(filepath.Join(mixinsDir, "helm/helm-runtime"))
}

// InitializePorterHome initializes the filesystem with the supporting files in the PORTER_HOME directory.
func (c *TestConfig) SetupIntegrationTest(home string) {
	c.SetHomeDir(home)
	// Copy bin dir contents to the home directory
	c.TestContext.AddTestDirectory(c.TestContext.FindBinDir(), home)
}
