package porter

import (
	"github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDependencyExecutioner_ExecuteBeforePrepare(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.TestContext.AddTestFile("../../build/testdata/bundles/wordpress/porter.yaml", "porter.yaml")

	err := p.LoadManifest()
	require.NoError(t, err)

	e := newDependencyExecutioner(p.Porter)

	// Try to call execute without prepare
	err = e.Execute()
	require.Error(t, err, "execute before prepare should return an error")
	assert.EqualError(t, err, "Prepare must be called before Execute")

	// Now make sure execute passes now that we have called execute
	opts := InstallOptions{}
	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err, "opts validate failed")
	err = e.Prepare(opts.BundleLifecycleOpts, func(args cnabprovider.ActionArguments) error {
		return nil
	})
	require.NoError(t, err, "prepare should have succeeded")
	err = e.Execute()
	require.NoError(t, err, "execute should not fail when we have called prepare")
}
