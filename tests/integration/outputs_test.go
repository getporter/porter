//go:build integration
// +build integration

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
	p.SetupIntegrationTest()
	ctx := context.Background()

	// Install a bundle with exec outputs
	bundleName := installExecOutputsBundle(p)
	defer CleanupCurrentBundle(p)

	// Verify that its file output was captured
	usersOutput, err := p.ReadBundleOutput(ctx, "users.json", bundleName, "")
	require.NoError(t, err, "could not read users output")
	assert.Equal(t, fmt.Sprintln(`{"users": ["sally"]}`), usersOutput, "expected the users output to be populated correctly")

	// Verify that its bundle level file output was captured
	opts := porter.OutputListOptions{}
	opts.Name = bundleName
	opts.Format = printer.FormatPlaintext
	displayOutputs, err := p.ListBundleOutputs(context.Background(), &opts)
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

	invokeExecOutputsBundle(p, "add-user")
	invokeExecOutputsBundle(p, "get-users")

	// Verify logs were captured as an output
	logs, err := p.ReadBundleOutput(ctx, cnab.OutputInvocationImageLogs, bundleName, "")
	require.NoError(t, err, "ListBundleOutputs failed")
	assert.Contains(t, logs, "executing get-users action from exec-outputs", "expected the logs to contain bundle output from the last action")

	// Verify that its jsonPath output was captured
	userOutput, err := p.ReadBundleOutput(ctx, "user-names", bundleName, "")
	require.NoError(t, err, "could not read user-names output")
	assert.Equal(t, `["sally","wei"]`, userOutput, "expected the user-names output to be populated correctly")

	invokeExecOutputsBundle(p, "test")

	// Verify that its regex output was captured
	testOutputs, err := p.ReadBundleOutput(ctx, "failed-tests", bundleName, "")
	require.NoError(t, err, "could not read failed-tests output")
	assert.Equal(t, "TestInstall\nTestUpgrade", testOutputs, "expected the failed-tests output to be populated correctly")
}

func CleanupCurrentBundle(p *porter.TestPorter) {
	// Uninstall the bundle
	uninstallOpts := porter.NewUninstallOptions()
	err := uninstallOpts.Validate(context.Background(), []string{}, p.Porter)
	assert.NoError(p.T(), err, "validation of uninstall opts failed for current bundle")

	err = p.UninstallBundle(context.Background(), uninstallOpts)
	assert.NoError(p.T(), err, "uninstall failed for current bundle")
}

func installExecOutputsBundle(p *porter.TestPorter) string {
	err := p.Create()
	require.NoError(p.T(), err)

	bundleName := p.AddTestBundleDir("testdata/bundles/exec-outputs", true)

	installOpts := porter.NewInstallOptions()
	err = installOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(p.T(), err)
	err = p.InstallBundle(context.Background(), installOpts)
	require.NoError(p.T(), err)

	return bundleName
}

func invokeExecOutputsBundle(p *porter.TestPorter, action string) {
	statusOpts := porter.NewInvokeOptions()
	statusOpts.Action = action
	err := statusOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(p.T(), err)
	err = p.InvokeBundle(context.Background(), statusOpts)
	require.NoError(p.T(), err, "invoke %s should have succeeded", action)
}

func TestStepLevelAndBundleLevelOutputs(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	p.SetupIntegrationTest()
	p.Debug = false

	p.AddTestBundleDir("testdata/bundles/outputs-example", true)

	// Install the bundle
	// A step-level output will be used during this action
	installOpts := porter.NewInstallOptions()
	err := installOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(t, err)
	err = p.InstallBundle(context.Background(), installOpts)
	require.NoError(t, err, "install should have succeeded")

	// Upgrade the bundle
	// A bundle-level output will be produced during this action
	upgradeOpts := porter.NewUpgradeOptions()
	err = upgradeOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(t, err)
	err = p.UpgradeBundle(context.Background(), upgradeOpts)
	require.NoError(t, err, "upgrade should have succeeded")

	// Uninstall the bundle
	// A bundle-level output will be used during this action
	uninstallOpts := porter.NewUninstallOptions()
	err = uninstallOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(t, err)
	err = p.UninstallBundle(context.Background(), uninstallOpts)
	require.NoError(t, err, "uninstall should have succeeded")
}
