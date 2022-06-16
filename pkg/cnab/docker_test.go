package cnab

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessedExtensions_GetDockerExtension(t *testing.T) {
	t.Parallel()

	t.Run("extension present", func(t *testing.T) {
		t.Parallel()

		ext := ProcessedExtensions{
			DockerExtensionKey: Docker{
				Privileged: true,
			},
		}

		dockerExt, dockerRequired, err := ext.GetDocker()
		require.NoError(t, err, "GetDocker failed")
		assert.True(t, dockerRequired, "docker should be a required extension")
		assert.Equal(t, Docker{Privileged: true}, dockerExt, "docker config was not populated properly")
	})

	t.Run("extension missing", func(t *testing.T) {
		t.Parallel()

		ext := ProcessedExtensions{}

		dockerExt, dockerRequired, err := ext.GetDocker()
		require.NoError(t, err, "GetDocker failed")
		assert.False(t, dockerRequired, "docker should NOT be a required extension")
		assert.Equal(t, Docker{}, dockerExt, "Docker config should default to empty when not required")
	})

	t.Run("extension invalid", func(t *testing.T) {
		t.Parallel()

		ext := ProcessedExtensions{
			DockerExtensionKey: map[string]string{"ponies": "are great"},
		}

		dockerExt, dockerRequired, err := ext.GetDocker()
		require.Error(t, err, "GetDocker should have failed")
		assert.True(t, dockerRequired, "docker should be a required extension")
		assert.Equal(t, Docker{}, dockerExt, "Docker config should default to empty")
	})
}

func TestSupportsDocker(t *testing.T) {
	t.Parallel()

	t.Run("supported", func(t *testing.T) {
		b := NewBundle(bundle.Bundle{
			RequiredExtensions: []string{DockerExtensionKey},
		})

		assert.True(t, b.SupportsDocker())
	})
	t.Run("unsupported", func(t *testing.T) {
		b := NewBundle(bundle.Bundle{})

		assert.False(t, b.SupportsDocker())
	})
}
