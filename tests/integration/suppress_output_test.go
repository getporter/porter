//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"get.porter.sh/porter/pkg/porter"
)

func TestSuppressOutput(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest()
	p.Debug = false

	bundleName := p.AddTestBundleDir("testdata/bundles/suppressed-output-example", true)

	// Install (Output suppressed)
	installOpts := porter.NewInstallOptions()
	err := installOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(context.Background(), installOpts)
	require.NoError(t, err)

	// Verify that the bundle output was captured (despite stdout/err of command being suppressed)
	bundleOutput, err := p.ReadBundleOutput("greeting", bundleName, "")
	require.NoError(t, err, "could not read config output")
	require.Equal(t, "Hello World!", bundleOutput, "expected the bundle output to be populated correctly")

	// Invoke - Log Error (Output suppressed)
	invokeOpts := porter.NewInvokeOptions()
	invokeOpts.Action = "log-error"
	err = invokeOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InvokeBundle(context.Background(), invokeOpts)
	require.NoError(t, err)

	// Uninstall
	uninstallOpts := porter.NewUninstallOptions()
	err = uninstallOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(t, err)

	err = p.UninstallBundle(context.Background(), uninstallOpts)
	require.NoError(t, err)

	gotCmdOutput := p.TestConfig.TestContext.GetOutput()

	require.NotContains(t, gotCmdOutput, "Hello World!", "expected command output to be suppressed from Install step")
	require.NotContains(t, gotCmdOutput, "Error!", "expected command output to be suppressed from Invoke step")
	require.Contains(t, gotCmdOutput, "Farewell World!", "expected command output to be present from Uninstall step")
}
