package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencyExecutioner_ExecuteBeforePrepare(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestDirectoryFromRoot("build/testdata/bundles/mysql", "/")
	err := p.LoadManifestFrom("/porter.yaml")
	require.NoError(t, err)

	e := newDependencyExecutioner(p.Porter, cnab.ActionInstall)

	// Try to call execute without prepare
	err = e.Execute()
	require.Error(t, err, "execute before prepare should return an error")
	assert.EqualError(t, err, "Prepare must be called before Execute")

	// Now make sure execute passes now that we have called execute
	opts := NewInstallOptions()
	opts.Driver = DebugDriver
	opts.File = "/porter.yaml"
	err = opts.Validate([]string{}, p.Porter)
	require.NoError(t, err, "opts validate failed")
	err = e.Prepare(opts)
	require.NoError(t, err, "prepare should have succeeded")
	err = e.Execute()
	require.NoError(t, err, "execute should not fail when we have called prepare")
}
