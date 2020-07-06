// +build integration

package tests

import (
	"fmt"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecOutputs(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	// Install a bundle with exec outputs
	installExecOutputsBundle(p)
	defer CleanupCurrentBundle(p)

	// Verify that its file output was captured
	configOutput, err := p.ReadBundleOutput("config", p.Manifest.Name)
	require.NoError(t, err, "could not read config output")
	assert.Equal(t, fmt.Sprintln(`{"user": "sally"}`), configOutput, "expected the config output to be populated correctly")

	// Verify that its bundle level file output was captured
	opts := porter.OutputListOptions{}
	opts.Name = p.Manifest.Name
	opts.Format = printer.FormatTable
	displayOutputs, err := p.ListBundleOutputs(&opts)
	require.NoError(t, err, "ListBundleOutputs failed")
	var kubeconfigOutput *porter.DisplayOutput
	for _, do := range displayOutputs {
		if do.Name == "kubeconfig" {
			kubeconfigOutput = &do
			break
		}
	}
	require.NotNil(t, kubeconfigOutput, "could not find kubeconfig output")
	assert.Equal(t, "file", kubeconfigOutput.Type)
	assert.Contains(t, kubeconfigOutput.Value, "apiVersion")

	invokeExecOutputsBundle(p, "status")

	// Verify that its jsonPath output was captured
	userOutput, err := p.ReadBundleOutput("user", p.Manifest.Name)
	require.NoError(t, err, "could not read user output")
	assert.Equal(t, "sally", userOutput, "expected the user output to be populated correctly")

	invokeExecOutputsBundle(p, "test")

	// Verify that its regex output was captured
	testOutputs, err := p.ReadBundleOutput("failed-tests", p.Manifest.Name)
	require.NoError(t, err, "could not read failed-tests output")
	assert.Equal(t, "TestInstall\nTestUpgrade", testOutputs, "expected the failed-tests output to be populated correctly")
}

func CleanupCurrentBundle(p *porter.TestPorter) {
	// Uninstall the bundle
	uninstallOpts := porter.UninstallOptions{}
	err := uninstallOpts.Validate([]string{}, p.Context)
	assert.NoError(p.T(), err, "validation of uninstall opts failed for current bundle")

	err = p.UninstallBundle(uninstallOpts)
	assert.NoError(p.T(), err, "uninstall failed for current bundle")
}

func installExecOutputsBundle(p *porter.TestPorter) {
	err := p.Create()
	require.NoError(p.T(), err)

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.TestDir, "../examples/exec-outputs"), ".")

	x := p.Context.FileSystem
	files, _ := x.ReadDir(".")
	fmt.Println(files)

	installOpts := porter.InstallOptions{}
	err = installOpts.Validate([]string{}, p.Context)
	require.NoError(p.T(), err)
	err = p.InstallBundle(installOpts)
	require.NoError(p.T(), err)
}

func invokeExecOutputsBundle(p *porter.TestPorter, action string) {
	statusOpts := porter.InvokeOptions{Action: action}
	err := statusOpts.Validate([]string{}, p.Context)
	require.NoError(p.T(), err)
	err = p.InvokeBundle(statusOpts)
	require.NoError(p.T(), err, "invoke %s should have succeeded", action)
}

func TestStepLevelAndBundleLevelOutputs(t *testing.T) {
	t.Skip("TODO: Implement parameter sources #1067")

	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.TestDir, "testdata/bundles/outputs-example"), ".")

	// Install the bundle
	// A step-level output will be used during this action
	installOpts := porter.InstallOptions{}
	err := installOpts.Validate([]string{}, p.Context)
	require.NoError(t, err)
	err = p.InstallBundle(installOpts)
	require.NoError(t, err, "install should have succeeded")

	// Upgrade the bundle
	// A bundle-level output will be produced during this action
	upgradeOpts := porter.UpgradeOptions{}
	err = upgradeOpts.Validate([]string{}, p.Context)
	require.NoError(t, err)
	err = p.UpgradeBundle(upgradeOpts)
	require.NoError(t, err, "upgrade should have succeeded")

	// Uninstall the bundle
	// A bundle-level output will be used during this action
	uninstallOpts := porter.UninstallOptions{}
	err = uninstallOpts.Validate([]string{}, p.Context)
	require.NoError(t, err)
	err = p.UninstallBundle(uninstallOpts)
	require.NoError(t, err, "uninstall should have succeeded")
}
