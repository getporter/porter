package extensions

import (
	"io/ioutil"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessedExtensions_GetParameterSourcesExtension(t *testing.T) {
	t.Run("extension present", func(t *testing.T) {
		var ps ParameterSources
		ps.SetParameterFromOutput("tfstate", "tfstate")
		processed := ProcessedExtensions{
			ParameterSourcesKey: ps,
		}

		ext, required, err := processed.GetParameterSources()
		require.NoError(t, err, "GetParameterSources failed")
		assert.True(t, required, "parameter-sources should be a required extension")
		assert.Equal(t, ps, ext, "parameter-sources was not populated properly")
	})

	t.Run("extension missing", func(t *testing.T) {
		processed := ProcessedExtensions{}

		ext, required, err := processed.GetParameterSources()
		require.NoError(t, err, "GetParameterSources failed")
		assert.False(t, required, "parameter-sources should NOT be a required extension")
		assert.Empty(t, ext, "parameter-sources should default to empty when not required")
	})

	t.Run("extension invalid", func(t *testing.T) {
		processed := ProcessedExtensions{
			ParameterSourcesKey: map[string]string{"ponies": "are great"},
		}

		ext, required, err := processed.GetParameterSources()
		require.Error(t, err, "GetParameterSources should have failed")
		assert.True(t, required, "parameter-sources should be a required extension")
		assert.Empty(t, ext, "parameter-sources should default to empty when not required")
	})
}

func TestReadParameterSourcesProperties(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/bundle.json")
	require.NoError(t, err, "cannot read bundle file")

	bun, err := bundle.Unmarshal(data)
	require.NoError(t, err, "could not unmarshal the bundle")
	assert.True(t, HasParameterSources(*bun))

	ps, err := ReadParameterSources(*bun)

	want := ParameterSources{}
	want.SetParameterFromOutput("tfstate", "tfstate")
	assert.Equal(t, want, ps)
}
