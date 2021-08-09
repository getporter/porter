package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_Run(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	// Mock the mixin test runner and verify that we are calling runtime mixins, e.g. exec-runtime and not exec
	mp := p.Mixins.(*mixin.TestMixinProvider)
	mp.RunAssertions = append(mp.RunAssertions, func(mixinCxt *context.Context, mixinName string, commandOpts pkgmgmt.CommandOptions) error {
		assert.Equal(t, "exec", mixinName, "expected to call the exec mixin")
		assert.True(t, commandOpts.Runtime, "the mixin command should be executed in runtime mode")
		assert.Equal(t, "install", commandOpts.Command, "should have executed the mixin's install command")
		return nil
	})
	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/cnab/bundle.json")
	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")
	p.FileSystem.Create("/root/.kube/config")

	opts := NewRunOptions(p.Config)
	opts.Action = cnab.ActionInstall
	opts.File = "porter.yaml"

	err := opts.Validate()
	require.NoError(t, err, "could not validate run options")

	err = p.Run(opts)
	assert.NoError(t, err, "run failed")
}

func TestPorter_defaultDebugToOff(t *testing.T) {
	p := New() // Don't use the test porter, it has debug on by default
	defer p.Close()

	opts := NewRunOptions(p.Config)

	err := opts.defaultDebug()
	require.NoError(t, err)
	assert.False(t, p.Config.Debug)
}

func TestPorter_defaultDebugUsesEnvVar(t *testing.T) {
	p := New() // Don't use the test porter, it has debug on by default
	defer p.Close()

	p.Setenv(config.EnvDEBUG, "true")

	opts := NewRunOptions(p.Config)

	err := opts.defaultDebug()
	require.NoError(t, err)

	assert.True(t, p.Config.Debug)
}
