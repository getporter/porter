package config

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
)

type TestConfig struct {
	*Config
	TestContext *portercontext.TestContext
	TestSpan    tracing.RootTraceLogger
}

// NewTestConfig initializes a configuration suitable for testing:
// * buffered output,
// * in-memory file system,
// * does not automatically load config from ambient environment.
func NewTestConfig(t *testing.T) *TestConfig {
	cxt := portercontext.NewTestContext(t)
	cfg := NewFor(cxt.Context)
	cfg.Data.Verbosity = "debug"
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
	c.FileSystem.Create(filepath.Join(mixinsDir, "testmixin/testmixin"))
	c.FileSystem.Create(filepath.Join(mixinsDir, "testmixin/runtimes/testmixin-runtime"))
}

// SetupIntegrationTest initializes the filesystem with the supporting files in
// a temp PORTER_HOME directory.
func (c *TestConfig) SetupIntegrationTest() (ctx context.Context, testDir string, homeDir string) {
	ctx = context.Background()
	testDir, homeDir = c.TestContext.UseFilesystem()
	c.SetHomeDir(homeDir)

	// Use the compiled porter binary in the test home directory,
	// and not the go test binary that is generated when we run integration tests.
	// This way when Porter calls back to itself, e.g. for internal plugins,
	// it is calling the normal porter binary.
	c.SetPorterPath(filepath.Join(homeDir, "porter"))

	// Copy bin dir contents to the home directory
	c.TestContext.AddTestDirectory(c.TestContext.FindBinDir(), homeDir, pkg.FileModeDirectory)

	// Check if telemetry should be enabled for the test
	if telemetryEnabled, _ := strconv.ParseBool(os.Getenv("PORTER_TEST_TELEMETRY_ENABLED")); telemetryEnabled {
		// Okay someone is listening, configure the tracer
		c.Data.Telemetry.Enabled = true
		c.Data.Telemetry.Insecure = true
		c.Data.Telemetry.Protocol = "grpc"
		c.ConfigureLogging(ctx, c.NewLogConfiguration())
		ctx, c.TestSpan = c.StartRootSpan(ctx, c.TestContext.T.Name())
	}

	return ctx, testDir, homeDir
}

func (c *TestConfig) Close() {
	if c.TestSpan != nil {
		c.TestSpan.Close()
		c.TestSpan = nil
	}
	c.TestContext.Close()
}
