//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigContextList verifies that `porter config context list` correctly
// marks the active context and lists all available contexts.
func TestConfigContextList(t *testing.T) {
	test, err := tester.NewTestWithConfig(t, "tests/testdata/config/config-multi-context.yaml")
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	stdout, _ := test.RequirePorter("config", "context", "list")
	assert.Contains(t, stdout, "* default", "default should be marked active (matches current-context in file)")
	assert.Contains(t, stdout, "  prod", "prod should be listed as an inactive context")
}

// TestConfigContextUse verifies that `porter config context use` updates the
// active context in the config file, and that a subsequent list reflects it.
func TestConfigContextUse(t *testing.T) {
	test, err := tester.NewTestWithConfig(t, "tests/testdata/config/config-multi-context.yaml")
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Switch to prod context
	test.RequirePorter("config", "context", "use", "prod")

	// Verify the update persists across process invocations
	stdout, _ := test.RequirePorter("config", "context", "list")
	assert.Contains(t, stdout, "* prod", "prod should be marked active after 'context use prod'")
	assert.Contains(t, stdout, "  default", "default should still be listed as inactive")
}

// multiContextYAML is a minimal multi-context config for real-filesystem tests.
const multiContextYAML = `schemaVersion: "2.0.0"
current-context: default
contexts:
  - name: default
    config:
      namespace: dev
  - name: prod
    config:
      namespace: prod
`

// multiContextTOML is a multi-context config in TOML format for format-agnostic tests.
const multiContextTOML = `schemaVersion = "2.0.0"
current-context = "prod"

[[contexts]]
name = "default"
[contexts.config]
namespace = "ns-default"

[[contexts]]
name = "prod"
[contexts.config]
namespace = "ns-prod"
`

// TestMultiContextConfig_ContextSelectsNamespace verifies that the --context
// flag loads the namespace defined in the selected context.
func TestMultiContextConfig_ContextSelectsNamespace(t *testing.T) {
	p := porter.NewTestPorter(t)
	ctx := p.SetupIntegrationTest()
	defer p.Close()

	home, _ := p.GetHomeDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(home, "config.yaml"),
		[]byte(multiContextYAML),
		pkg.FileModeWritable,
	))

	p.Config.DataLoader = config.LoadFromFilesystem()

	// Default context (no --context flag) → namespace "dev"
	p.Config.ContextName = ""
	_, err := p.Config.Load(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, "dev", p.Config.Data.Namespace, "default context should set namespace to 'dev'")

	// --context prod → namespace "prod"
	p.Config.ContextName = "prod"
	_, err = p.Config.Load(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, "prod", p.Config.Data.Namespace, "--context prod should set namespace to 'prod'")
}

// legacyConfigYAML is a flat (pre-2.0.0) config file with no schemaVersion.
const legacyConfigYAML = `namespace: legacy
`

// TestLegacyConfig_StillWorks verifies that the old flat config format is
// still loaded correctly after adding multi-context support.
func TestLegacyConfig_StillWorks(t *testing.T) {
	p := porter.NewTestPorter(t)
	ctx := p.SetupIntegrationTest()
	defer p.Close()

	home, _ := p.GetHomeDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(home, "config.yaml"),
		[]byte(legacyConfigYAML),
		pkg.FileModeWritable,
	))

	p.Config.DataLoader = config.LoadFromFilesystem()

	// Flat config loads namespace directly
	p.Config.ContextName = ""
	_, err := p.Config.Load(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, "legacy", p.Config.Data.Namespace, "flat config should load namespace from the top level")

	// --context flag must fail with flat config
	p.Config.ContextName = "prod"
	_, err = p.Config.Load(ctx, nil)
	require.ErrorContains(t, err, "--context", "--context flag should be rejected for legacy flat configs")
}

// TestMultiContextConfig_TOML verifies that context listing works with
// TOML-formatted config files, exercising the format-agnostic viper approach.
func TestMultiContextConfig_TOML(t *testing.T) {
	p := porter.NewTestPorter(t)
	ctx := p.SetupIntegrationTest()
	defer p.Close()

	home, _ := p.GetHomeDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(home, "config.toml"),
		[]byte(multiContextTOML),
		pkg.FileModeWritable,
	))

	err := p.ConfigContextList(ctx)
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, output, "  default", "default should be listed as an inactive context")
	assert.Contains(t, output, "* prod", "prod should be marked active (matches current-context in file)")
}
