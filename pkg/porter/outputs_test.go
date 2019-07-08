package porter

import (
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/printer"
	"github.com/stretchr/testify/require"
)

func TestPorter_ListOutputs(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	homeDir, err := p.TestConfig.GetHomeDir()
	require.NoError(t, err)

	p.TestConfig.TestContext.AddTestDirectory("testdata/outputs", filepath.Join(homeDir, "outputs"))

	opts := OutputListOptions{
		Bundle: "test-bundle",
		Format: printer.FormatTable,
	}
	err = p.ListBundleOutputs(opts)
	require.NoError(t, err)

	wantContains := []string{
		"NAME   MODIFIED",
		"foo    now",
		"bar    now",
	}

	gotOutput := p.TestConfig.TestContext.GetOutput()
	for _, want := range wantContains {
		require.Contains(t, gotOutput, want)
	}
}

func TestPorter_ShowOutput(t *testing.T) {
	// TODO
}
