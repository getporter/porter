package config

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/context"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	*Config
	TestContext *context.TestContext
}

// NewTestConfig initializes a configuration suitable for testing, with the output buffered, and an in-memory file system.
func NewTestConfig(t *testing.T) *TestConfig {
	cxt := context.NewTestContext(t)
	cfg := New()
	cfg.Context = cxt.Context
	tc := &TestConfig{
		Config:      cfg,
		TestContext: cxt,
	}
	tc.SetupUnitTest()
	return tc
}

func (c *TestConfig) T() *testing.T {
	return c.TestContext.T
}

// SetupUnitTest initializes the unit test filesystem with the supporting files in the PORTER_HOME directory.
func (c *TestConfig) SetupUnitTest() {
	// Set up the test porter home directory using an absolute path based on the OS filesystem
	// e.g. C:\.porter on Windows, or /.porter otherwise.
	rootDir, err := filepath.Abs("/")
	require.NoError(c.T(), err, "could not determine the filesystem root")
	home := filepath.Join(rootDir, ".porter")
	c.SetHomeDir(home)

	// Fake out the porter home directory
	c.FileSystem.Create(filepath.Join(home, "porter"))
	c.FileSystem.Create(filepath.Join(home, "runtimes", "porter-runtime"))

	mixinsDir := filepath.Join(home, "mixins")
	c.FileSystem.Create(filepath.Join(mixinsDir, "exec/exec"))
	c.FileSystem.Create(filepath.Join(mixinsDir, "exec/runtimes/exec-runtime"))
	c.FileSystem.Create(filepath.Join(mixinsDir, "helm/helm"))
	c.FileSystem.Create(filepath.Join(mixinsDir, "helm/runtimes/helm-runtime"))
}

// SetupIntegrationTest initializes the filesystem with the supporting files in
// a temp PORTER_HOME directory.
func (c *TestConfig) SetupIntegrationTest() (testDir string, homeDir string) {
	testDir, homeDir = c.TestContext.UseFilesystem()
	c.SetHomeDir(homeDir)

	// Use the compiled porter binary in the test home directory,
	// and not the go test binary that is generated when we run integration tests.
	// This way when Porter calls back to itself, e.g. for internal plugins,
	// it is calling the normal porter binary.
	c.SetPorterPath(filepath.Join(homeDir, "porter"))

	// Copy bin dir contents to the home directory
	c.TestContext.AddTestDirectory(c.TestContext.FindBinDir(), homeDir)

	// Remove any rando stuff copied from the dev bin, you won't find this in CI but a local dev run may have it
	// Not checking for an error, since the files won't be there on CI
	c.FileSystem.RemoveAll(filepath.Join(homeDir, "installations"))
	c.FileSystem.RemoveAll(filepath.Join(homeDir, "claims"))
	c.FileSystem.RemoveAll(filepath.Join(homeDir, "results"))
	c.FileSystem.RemoveAll(filepath.Join(homeDir, "outputs"))
	c.FileSystem.Remove(filepath.Join(homeDir, "schema.json"))

	return testDir, homeDir
}
