package porter

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPorter_listBundleOutputs(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	homeDir, err := p.TestConfig.GetHomeDir()
	require.NoError(t, err)

	p.TestConfig.TestContext.AddTestDirectory("testdata/outputs", filepath.Join(homeDir, "outputs"))

	got, err := p.listBundleOutputs("test-bundle")
	require.NoError(t, err)

	want := &Outputs{
		{
			Name:      "foo",
			Type:      "string",
			Value:     "foo-value",
			Sensitive: true,
		},
		{
			Name:      "bar",
			Type:      "string",
			Value:     "bar-value",
			Sensitive: false,
		},
	}

	require.Equal(t, want, got)
}
