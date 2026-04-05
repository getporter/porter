//go:build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvokeCustomAction(t *testing.T) {
	// this is sentinel output that is only output by the porter runtime when in debug mode
	const runtimeDebugOutputCheck = "=== Step Template ==="

	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	// Install a bundle with a custom action defined
	err := p.Create()
	require.NoError(t, err)

	bundleName := p.AddTestBundleDir("testdata/bundles/bundle-with-custom-action", true)

	installOpts := porter.NewInstallOptions()
	// explicitly do not set --debug for install
	err = installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)
	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err)

	// Make sure that when --debug is not passed, we do not output porter runtimes debug lines
	gotErr := p.TestConfig.TestContext.GetError()
	require.NotContains(t, gotErr, runtimeDebugOutputCheck, "expected no debug output from the porter runtime since --debug was not passed")

	// Invoke the custom action
	invokeOpts := porter.NewInvokeOptions()
	invokeOpts.DebugMode = true
	invokeOpts.Action = "zombies"
	err = invokeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)
	err = p.InvokeBundle(ctx, invokeOpts)
	require.NoError(t, err, "invoke should have succeeded")

	gotOutput := p.TestConfig.TestContext.GetOutput()
	tests.RequireOutputContains(t, gotOutput, "oh noes my brains", "invoke should have printed a cry for halp")

	// Check that debug output from the porter runtime was printed by the bundle and porter collected it
	// This checks that the PORTER_DEBUG parameter is being properly passed to a bundle when run with porter invoke --debug
	gotStderr := p.TestConfig.TestContext.GetOutput()
	tests.RequireOutputContains(t, gotStderr, runtimeDebugOutputCheck, "expected debug output from the porter runtime to be output by the bundle")

	// Verify that the custom action was recorded properly
	i, err := p.Installations.GetInstallation(ctx, "", bundleName)
	require.NoError(t, err, "could not fetch installation")
	c, err := p.Installations.GetLastRun(ctx, i.Namespace, i.Name)
	require.NoError(t, err, "GetLastClaim failed")
	assert.Equal(t, "zombies", c.Action, "the custom action wasn't recorded in the installation")
}

func TestInvoke_ModifiesFalse_DoesNotPersistState(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	err := p.Create()
	require.NoError(t, err)

	bundleName := p.AddTestBundleDir("testdata/bundles/bundle-with-modifies-false-action", true)

	// Install to establish initial porter-state
	installOpts := porter.NewInstallOptions()
	err = installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)
	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err, "install should have succeeded")

	// Capture the porter-state output written during install
	stateAfterInstall, err := p.Installations.GetLastOutput(ctx, "", bundleName, "porter-state")
	require.NoError(t, err, "porter-state output should exist after install")

	// Invoke the modifies:false dry-run action
	invokeOpts := porter.NewInvokeOptions()
	invokeOpts.Action = "dry-run"
	err = invokeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)
	err = p.InvokeBundle(ctx, invokeOpts)
	require.NoError(t, err, "dry-run invoke should have succeeded")

	// porter-state must not have been updated by the dry-run
	stateAfterDryRun, err := p.Installations.GetLastOutput(ctx, "", bundleName, "porter-state")
	require.NoError(t, err, "porter-state output should still exist after dry-run")
	assert.Equal(t, stateAfterInstall.ResultID, stateAfterDryRun.ResultID,
		"porter-state should not have been updated by a modifies:false action")

	// The dry-run run record must exist
	lastRun, err := p.Installations.GetLastRun(ctx, "", bundleName)
	require.NoError(t, err, "GetLastRun failed")
	assert.Equal(t, "dry-run", lastRun.Action, "dry-run should have been recorded")

	// The user-defined output from dry-run must have been saved
	dryRunOutput, err := p.Installations.GetOutput(ctx, lastRun.ID, "dry-run-result")
	require.NoError(t, err, "dry-run-result output should have been saved")
	assert.Contains(t, string(dryRunOutput.Value), "dry-run-ok")
}
