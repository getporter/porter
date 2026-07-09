//go:build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/require"
)

// TestInstall_lockedWhileRunIncomplete verifies that porter refuses to start
// a new run for an installation that already has an incomplete run when the
// dependencies-v2 feature flag is enabled, and that --force-run bypasses it.
func TestInstall_lockedWhileRunIncomplete(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()
	p.TestConfig.SetExperimentalFlags(experimental.FlagDependenciesV2)

	bundleName := p.AddTestBundleDir("testdata/bundles/bundle-with-custom-action", true)

	// Simulate a run that is still in progress for this installation
	blockingRun := p.TestInstallations.CreateRun(storage.NewRun("", bundleName), p.TestInstallations.SetMutableRunValues)
	p.TestInstallations.CreateResult(blockingRun.NewResult(cnab.StatusRunning))

	installOpts := porter.NewInstallOptions()
	require.NoError(t, installOpts.Validate(ctx, []string{}, p.Porter))

	err := p.InstallBundle(ctx, installOpts)
	require.Error(t, err, "install should be blocked while another run is incomplete")
	require.ErrorContains(t, err, "incomplete run")

	installOpts.ForceRun = true
	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err, "--force-run should bypass the lock")
}
