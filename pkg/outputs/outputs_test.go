package outputs

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/deislabs/porter/pkg/config"
)

func TestPorter_ReadBundleOutput(t *testing.T) {
	c := config.NewTestConfig(t)

	homeDir, err := c.GetHomeDir()
	require.NoError(t, err)

	c.TestContext.AddTestDirectory("../porter/testdata/outputs", filepath.Join(homeDir, "outputs"))

	got, err := ReadBundleOutput(c.Config, "foo", "test-bundle")
	require.NoError(t, err)

	want := &Output{
		Name:      "foo",
		Type:      "string",
		Value:     "foo-value",
		Sensitive: true,
	}

	require.Equal(t, want, got)
}
