// +build integration

package tests

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"get.porter.sh/porter/pkg/porter"
)

func TestSuppressOutput(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	// Currently, the default docker driver prints directly to stdout instead of to the given writer
	// Hence, the need to swap out stdout/stderr just for this test, to capture output under test
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.TestDir, "testdata/bundles/suppressed-output-example"), ".")

	// Install (Output suppressed)
	installOpts := porter.InstallOptions{}
	err := installOpts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.InstallBundle(installOpts)
	require.NoError(t, err)

	// Verify that the bundle output was captured (despite stdout/err of command being suppressed)
	bundleOutput, err := p.ReadBundleOutput("greeting", p.Manifest.Name)
	require.NoError(t, err, "could not read config output")
	require.Equal(t, "Hello World!", bundleOutput, "expected the bundle output to be populated correctly")

	// Invoke - Log Error (Output suppressed)
	invokeOpts := porter.InvokeOptions{Action: "log-error"}
	err = invokeOpts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.InvokeBundle(invokeOpts)
	require.NoError(t, err)

	// Uninstall
	uninstallOpts := porter.UninstallOptions{}
	err = uninstallOpts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.UninstallBundle(uninstallOpts)
	require.NoError(t, err)

	// Copy the output in a separate goroutine so printing can't block indefinitely
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// Read our faked cmd output
	w.Close()
	gotCmdOutput := <-outC

	require.NotContains(t, gotCmdOutput, "Hello World!", "expected command output to be suppressed from Install step")
	require.NotContains(t, gotCmdOutput, "Error!", "expected command output to be suppressed from Invoke step")
	require.Contains(t, gotCmdOutput, "Farewell World!", "expected command output to be present from Uninstall step")
}
