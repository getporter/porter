package porter

import (
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/printer"
	"github.com/stretchr/testify/require"
)

func TestPorter_ShowBundle(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	homeDir, err := p.TestConfig.GetHomeDir()
	require.NoError(t, err)

	p.TestConfig.TestContext.AddTestDirectory("testdata/outputs", filepath.Join(homeDir, "outputs"))

	opts := ShowOptions{
		Name:   "test-bundle",
		Format: printer.FormatTable,
	}
	err = p.ShowBundle(opts)
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
