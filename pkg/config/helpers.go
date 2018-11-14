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
	cxt := context.NewTestContext(t)
	c := &TestConfig{
		Config: &Config{
			Context: cxt.Context,
		},
		TestContext: cxt,
	}
	return c
}

// InitializePorterHome initializes the test filesystem with the supporting files in the PORTER_HOME directory.
func (c *TestConfig) SetupPorterHome() {
	// Set up the test porter home directory
	home := "/root/.porter"
	os.Setenv(EnvHOME, home)

	// Copy bin dir contents to the home directory
	c.TestContext.AddTestDirectory("../../bin/", home)
}
