package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessedExtensions_GetParameterSourcesExtension(t *testing.T) {
	t.Run("extension present", func(t *testing.T) {
		var ps ParameterSources
		ps.SetParameterFromOutput("tfstate", "tfstate")
		processed := ProcessedExtensions{
			ParameterSourcesExtensionKey: ps,
		}

		ext, required, err := processed.GetParameterSourcesExtension()
		require.NoError(t, err, "GetParameterSourcesExtension failed")
		assert.True(t, required, "parameter-sources should be a required extension")
		assert.Equal(t, ps, ext, "parameter-sources was not populated properly")
	})

	t.Run("extension missing", func(t *testing.T) {
		processed := ProcessedExtensions{}

		ext, required, err := processed.GetParameterSourcesExtension()
		require.NoError(t, err, "GetParameterSourcesExtension failed")
		assert.False(t, required, "parameter-sources should NOT be a required extension")
		assert.Empty(t, ext, "parameter-sources should default to empty when not required")
	})

	t.Run("extension invalid", func(t *testing.T) {
		processed := ProcessedExtensions{
			ParameterSourcesExtensionKey: map[string]string{"ponies": "are great"},
		}

		ext, required, err := processed.GetParameterSourcesExtension()
		require.Error(t, err, "GetParameterSourcesExtension should have failed")
		assert.True(t, required, "parameter-sources should be a required extension")
		assert.Empty(t, ext, "parameter-sources should default to empty when not required")
	})
}
