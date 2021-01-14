package extensions

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
)

func TestProcessedExtensions_FileParameterSupport(t *testing.T) {
	t.Parallel()

	t.Run("extension present", func(t *testing.T) {
		t.Parallel()

		ext := ProcessedExtensions{
			FileParameterExtensionKey: nil,
		}

		supported := ext.FileParameterSupport()
		assert.True(t, supported, "file parameters should be supported")
	})

	t.Run("extension missing", func(t *testing.T) {
		t.Parallel()

		ext := ProcessedExtensions{}

		supported := ext.FileParameterSupport()
		assert.False(t, supported, "file parameters should not be supported")
	})
}

func TestSupportsFileParameters(t *testing.T) {
	t.Parallel()

	t.Run("supported", func(t *testing.T) {
		b := bundle.Bundle{
			RequiredExtensions: []string{FileParameterExtensionKey},
		}

		assert.True(t, SupportsFileParameters(b))
	})
	t.Run("unsupported", func(t *testing.T) {
		b := bundle.Bundle{}

		assert.False(t, SupportsFileParameters(b))
	})
}
