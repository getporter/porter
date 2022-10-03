//go:build integration
// +build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/require"
)

func TestSuppressOutput(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	bundleName := p.AddTestBundleDir("testdata/bundles/suppressed-output-example", true)

	// Install (Output suppressed)
	installOpts := porter.NewInstallOptions()
	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err)

	// Verify that the bundle output was captured (despite stdout/err of command being suppressed)
	bundleOutput, err := p.ReadBundleOutput(ctx, "greeting", bundleName, "")
	require.NoError(t, err, "could not read config output")
	require.Equal(t, "Hello World!", bundleOutput, "expected the bundle output to be populated correctly")

	// Invoke - Log Error (Output suppressed)
	invokeOpts := porter.NewInvokeOptions()
	invokeOpts.Action = "log-error"
	err = invokeOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InvokeBundle(ctx, invokeOpts)
	require.NoError(t, err)

	// Uninstall
	uninstallOpts := porter.NewUninstallOptions()
	err = uninstallOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.UninstallBundle(ctx, uninstallOpts)
	require.NoError(t, err)

	gotCmdOutput := p.TestConfig.TestContext.GetOutput()

	require.NotContains(t, gotCmdOutput, "Hello World!", "expected command output to be suppressed from Install step")
	require.NotContains(t, gotCmdOutput, "Error!", "expected command output to be suppressed from Invoke step")
	require.Contains(t, gotCmdOutput, "Farewell World!", "expected command output to be present from Uninstall step")
}
