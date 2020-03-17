package porter

import (
	"os"
	"testing"

	"get.porter.sh/porter/pkg/mixin"

	"get.porter.sh/porter/pkg/pkgmgmt"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_Run(t *testing.T) {
	p := NewTestPorter(t)

	// Mock the mixin test runner and verify that we are calling runtime mixins, e.g. exec-runtime and not exec
	mp := p.Mixins.(*mixin.TestMixinProvider)
	mp.RunAssertions = append(mp.RunAssertions, func(mixinCxt *context.Context, mixinName string, commandOpts pkgmgmt.CommandOptions) {
		assert.Equal(t, "exec", mixinName, "expected to call the exec mixin")
		assert.True(t, commandOpts.Runtime, "the mixin command should be executed in runtime mode")
		assert.Equal(t, "install", commandOpts.Command, "should have executed the mixin's install command")
	})
	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/cnab/bundle.json")
	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")
	p.TestConfig.TestContext.FileSystem.Create("/root/.kube/config")

	opts := NewRunOptions(p.Config)
	opts.Action = string(manifest.ActionInstall)
	opts.File = "porter.yaml"

	err := opts.Validate()
	require.NoError(t, err, "could not validate run options")

	err = p.Run(opts)
	assert.NoError(t, err, "run failed")
}

func TestPorter_defaultDebugToOff(t *testing.T) {
	p := New() // Don't use the test porter, it has debug on by default
	opts := NewRunOptions(p.Config)

	err := opts.defaultDebug()
	require.NoError(t, err)
	assert.False(t, p.Config.Debug)
}

func TestPorter_defaultDebugUsesEnvVar(t *testing.T) {
	os.Setenv(config.EnvDEBUG, "true")
	defer os.Unsetenv(config.EnvDEBUG)

	p := New() // Don't use the test porter, it has debug on by default
	opts := NewRunOptions(p.Config)

	err := opts.defaultDebug()
	require.NoError(t, err)

	assert.True(t, p.Config.Debug)
}

func TestPorter_defaultDindToOff(t *testing.T) {
	p := New()
	opts := NewRunOptions(p.Config)

	err := opts.defaultDind()
	require.NoError(t, err)
	assert.False(t, p.Config.Dind)
}

func TestPorter_defaultDindUsesEnvVar(t *testing.T) {
	os.Setenv(config.EnvDIND, "true")
	defer os.Unsetenv(config.EnvDIND)

	p := New() // Don't use the test porter, it has debug on by default
	opts := NewRunOptions(p.Config)

	err := opts.defaultDind()
	require.NoError(t, err)

	assert.True(t, p.Config.Dind)
}