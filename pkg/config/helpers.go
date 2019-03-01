package config

import (
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/context"
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
	os.Setenv(EnvHOME, home)

	// Copy bin dir contents to the home directory
	c.TestContext.AddTestDirectory(c.TestContext.FindBinDir(), home)
}
