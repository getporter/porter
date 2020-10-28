package query

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestGenerator_BuildInput(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/porter.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	g := ManifestGenerator{Manifest: m}

	t.Run("no config", func(t *testing.T) {
		input := g.buildInputForMixin("exec")

		assert.Nil(t, input.Config, "exec mixin should have no config")
		assert.Len(t, input.Actions, 4, "expected 4 actions")

		require.Contains(t, input.Actions, "install")
		assert.Len(t, input.Actions["install"], 2, "expected 2 exec install steps")

		require.Contains(t, input.Actions, "upgrade")
		assert.Len(t, input.Actions["upgrade"], 1, "expected 1 exec upgrade steps")

		require.Contains(t, input.Actions, "uninstall")
		assert.Len(t, input.Actions["uninstall"], 1, "expected 1 exec uninstall steps")

		require.Contains(t, input.Actions, "status")
		assert.Len(t, input.Actions["status"], 1, "expected 1 exec status steps")
	})

	t.Run("with config", func(t *testing.T) {
		input := g.buildInputForMixin("az")
		assert.Equal(t, map[interface{}]interface{}{"extensions": []interface{}{"iot"}}, input.Config, "az mixin should have config")
	})
}
