//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecOutputs(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	// Install a bundle with exec outputs
	bundleName := installExecOutputsBundle(ctx, p)
	defer CleanupCurrentBundle(ctx, p)

	// Verify that its file output was captured
	usersOutput, err := p.ReadBundleOutput(ctx, "users.json", bundleName, "")
	require.NoError(t, err, "could not read users output")
	assert.Equal(t, fmt.Sprintln(`{"users": ["sally"]}`), usersOutput, "expected the users output to be populated correctly")

	// Verify that its bundle level file output was captured
	opts := porter.OutputListOptions{}
	opts.Name = bundleName
	opts.Format = printer.FormatPlaintext
	displayOutputs, err := p.ListBundleOutputs(ctx, &opts)
	require.NoError(t, err, "ListBundleOutputs failed")
	var kubeconfigOutput *porter.DisplayValue
	for _, do := range displayOutputs {
		if do.Name == "kubeconfig" {
			kubeconfigOutput = &do
			break
		}
	}
	require.NotNil(t, kubeconfigOutput, "could not find kubeconfig output")
	assert.Equal(t, "file", kubeconfigOutput.Type)
	assert.Contains(t, kubeconfigOutput.Value, "apiVersion")

	invokeExecOutputsBundle(ctx, p, "add-user")
	invokeExecOutputsBundle(ctx, p, "get-users")

	// Verify logs were captured as an output
	logs, err := p.ReadBundleOutput(ctx, cnab.OutputInvocationImageLogs, bundleName, "")
	require.NoError(t, err, "ListBundleOutputs failed")
	assert.Contains(t, logs, "executing get-users action from exec-outputs", "expected the logs to contain bundle output from the last action")

	// Verify that its jsonPath output was captured
	userOutput, err := p.ReadBundleOutput(ctx, "user-names", bundleName, "")
	require.NoError(t, err, "could not read user-names output")
	assert.Equal(t, `["sally","wei"]`, userOutput, "expected the user-names output to be populated correctly")

	invokeExecOutputsBundle(ctx, p, "test")

	// Verify that its regex output was captured
	testOutputs, err := p.ReadBundleOutput(ctx, "failed-tests", bundleName, "")
	require.NoError(t, err, "could not read failed-tests output")
	assert.Equal(t, "TestInstall\nTestUpgrade", testOutputs, "expected the failed-tests output to be populated correctly")

	// Upgrade the bundle, producing a bundle-level output
	upgradeExecOutputsBundle(ctx, p)

	// Verify that the bundle-level output produced during upgrade was captured
	upgradedUserOutput, err := p.ReadBundleOutput(ctx, "upgraded-user", bundleName, "")
	require.NoError(t, err, "could not read upgraded-user output")
	assert.Equal(t, "wei-upgraded", upgradedUserOutput, "expected the upgraded-user output to be populated correctly")

	// The deferred CleanupCurrentBundle call uninstalls the bundle, whose
	// uninstall action asserts that the bundle-level output produced during
	// upgrade is still readable in a later action.
}

func CleanupCurrentBundle(ctx context.Context, p *porter.TestPorter) {
	// Uninstall the bundle
	uninstallOpts := porter.NewUninstallOptions()
	err := uninstallOpts.Validate(ctx, []string{}, p.Porter)
	assert.NoError(p.T(), err, "validation of uninstall opts failed for current bundle")

	err = p.UninstallBundle(ctx, uninstallOpts)
	assert.NoError(p.T(), err, "uninstall failed for current bundle")
}

func installExecOutputsBundle(ctx context.Context, p *porter.TestPorter) string {
	err := p.Create()
	require.NoError(p.T(), err)

	bundleName := p.AddTestBundleDir("testdata/bundles/exec-outputs", true)

	installOpts := porter.NewInstallOptions()
	err = installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err)
	err = p.InstallBundle(ctx, installOpts)
	require.NoError(p.T(), err)

	return bundleName
}

func invokeExecOutputsBundle(ctx context.Context, p *porter.TestPorter, action string) {
	statusOpts := porter.NewInvokeOptions()
	statusOpts.Action = action
	err := statusOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err)
	err = p.InvokeBundle(ctx, statusOpts)
	require.NoError(p.T(), err, "invoke %s should have succeeded", action)
}

func upgradeExecOutputsBundle(ctx context.Context, p *porter.TestPorter) {
	upgradeOpts := porter.NewUpgradeOptions()
	err := upgradeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err)
	err = p.UpgradeBundle(ctx, upgradeOpts)
	require.NoError(p.T(), err, "upgrade should have succeeded")
}
