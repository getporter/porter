//go:build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/require"
)

// TestExecLocalScript verifies that the exec mixin can execute local scripts
// without requiring a ./ prefix (fixes issue #2420)
func TestExecLocalScript(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	err := p.Create()
	require.NoError(t, err)

	p.AddTestBundleDir("testdata/bundles/exec-local-script", true)

	installOpts := porter.NewInstallOptions()
	err = installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)
	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err, "local script should be executable without ./ prefix")
}

// TestExecLocalScriptWithPrefix verifies that existing behavior with ./ prefix still works
func TestExecLocalScriptWithPrefix(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	err := p.Create()
	require.NoError(t, err)

	p.AddTestBundleDir("testdata/bundles/exec-local-script-with-prefix", true)

	installOpts := porter.NewInstallOptions()
	err = installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)
	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err, "local script with ./ prefix should still work")
}
