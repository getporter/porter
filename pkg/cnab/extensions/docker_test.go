package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessedExtensions_GetDockerExtension(t *testing.T) {
	t.Run("extension present", func(t *testing.T) {
		ext := ProcessedExtensions{
			DockerExtensionKey: Docker{
				Privileged: true,
			},
		}

		dockerExt, dockerRequired, err := ext.GetDockerExtension()
		require.NoError(t, err, "GetDockerExtension failed")
		assert.True(t, dockerRequired, "docker should be a required extension")
		assert.Equal(t, Docker{Privileged: true}, dockerExt, "docker config was not populated properly")
	})

	t.Run("extension missing", func(t *testing.T) {
		ext := ProcessedExtensions{}

		dockerExt, dockerRequired, err := ext.GetDockerExtension()
		require.NoError(t, err, "GetDockerExtension failed")
		assert.False(t, dockerRequired, "docker should NOT be a required extension")
		assert.Equal(t, Docker{}, dockerExt, "Docker config should default to empty when not required")
	})

	t.Run("extension invalid", func(t *testing.T) {
		ext := ProcessedExtensions{
			DockerExtensionKey: map[string]string{"ponies": "are great"},
		}

		dockerExt, dockerRequired, err := ext.GetDockerExtension()
		require.Error(t, err, "GetDockerExtension should have failed")
		assert.True(t, dockerRequired, "docker should be a required extension")
		assert.Equal(t, Docker{}, dockerExt, "Docker config should default to empty")
	})
}
