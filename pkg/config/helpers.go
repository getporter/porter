package config

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/portercontext"
)

type TestConfig struct {
	*Config
	TestContext *portercontext.TestContext
}

// NewTestConfig initializes a configuration suitable for testing:
// * buffered output,
// * in-memory file system,
// * does not automatically load config from ambient environment.
func NewTestConfig(t *testing.T) *TestConfig {
	cxt := portercontext.NewTestContext(t)
	cfg := New()
	cfg.Context = cxt.Context
	cfg.DataLoader = NoopDataLoader
	tc := &TestConfig{
		Config:      cfg,
		TestContext: cxt,
	}
	tc.SetupUnitTest()
	return tc
}

// SetupUnitTest initializes the unit test filesystem with the supporting files in the PORTER_HOME directory.
func (c *TestConfig) SetupUnitTest() {
	// Set up the test porter home directory
	home := "/home/myuser/.porter"
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
	c.TestContext.AddTestDirectory(c.TestContext.FindBinDir(), homeDir, pkg.FileModeDirectory)

	return testDir, homeDir
}
