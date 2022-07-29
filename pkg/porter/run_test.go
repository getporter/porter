package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_Run(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	// Mock the mixin test runner and verify that we are calling runtime mixins, e.g. exec-runtime and not exec
	mp := p.Mixins.(*mixin.TestMixinProvider)
	mp.RunAssertions = append(mp.RunAssertions, func(mixinCxt *portercontext.Context, mixinName string, commandOpts pkgmgmt.CommandOptions) error {
		assert.Equal(t, "exec", mixinName, "expected to call the exec mixin")
		assert.True(t, commandOpts.Runtime, "the mixin command should be executed in runtime mode")
		assert.Equal(t, "install", commandOpts.Command, "should have executed the mixin's install command")
		return nil
	})
	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/cnab/bundle.json")
	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")

	// Change the schemaVersion to validate that the runtime ignores this field and runs regardless
	e := yaml.NewEditor(p.Context)
	require.NoError(t, e.ReadFile("porter.yaml"))
	require.NoError(t, e.SetValue("schemaVersion", ""))
	require.NoError(t, e.WriteFile("porter.yaml"))

	p.FileSystem.Create("/home/nonroot/.kube/config")

	opts := NewRunOptions(p.Config)
	opts.Action = cnab.ActionInstall
	opts.File = "porter.yaml"

	err := opts.Validate()
	require.NoError(t, err, "could not validate run options")

	err = p.Run(context.Background(), opts)
	assert.NoError(t, err, "run failed")
}

func TestPorter_defaultDebugToOff(t *testing.T) {
	p := New() // Don't use the test porter, it has debug on by default
	defer p.Close()

	opts := NewRunOptions(p.Config)

	err := opts.defaultDebug()
	require.NoError(t, err)
	assert.False(t, opts.DebugMode)
}

func TestPorter_defaultDebugUsesEnvVar(t *testing.T) {
	p := New() // Don't use the test porter, it has debug on by default
	defer p.Close()

	p.Setenv(config.EnvDEBUG, "true")

	opts := NewRunOptions(p.Config)

	err := opts.defaultDebug()
	require.NoError(t, err)

	assert.True(t, opts.DebugMode)
}
