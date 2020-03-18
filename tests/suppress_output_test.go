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
	defer func() {
		os.Stdout = stdout
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w

	p.TestConfig.TestContext.AddTestFile(filepath.Join(p.TestDir, "testdata/bundle-with-suppressed-output.yaml"), "porter.yaml")

	installOpts := porter.InstallOptions{}
	err := installOpts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.InstallBundle(installOpts)
	require.NoError(t, err)

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

	// Read our faked stdout
	w.Close()
	gotstdout := <-outC

	require.NotContains(t, gotstdout, "Hello World", "expected output to be suppressed from Install step")
	require.Contains(t, gotstdout, "Goodbye World", "expected output to be present from Uninstall step")
}
