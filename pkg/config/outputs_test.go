package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPorter_ReadBundleOutput(t *testing.T) {
	c := NewTestConfig(t)

	homeDir, err := c.GetHomeDir()
	require.NoError(t, err)

	c.TestContext.AddTestDirectory("../porter/testdata/outputs", filepath.Join(homeDir, "outputs"))

	got, err := c.ReadBundleOutput("foo", "test-bundle")
	require.NoError(t, err)

	want := &Output{
		Name:      "foo",
		Type:      "string",
		Value:     "foo-value",
		Sensitive: true,
	}

	require.Equal(t, want, got)
}
