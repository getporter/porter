package cnab

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundleReference_CalculateInstallerImageName(t *testing.T) {
	bunRef := MustParseOCIReference("example.com/mybuns:v1.0.0")

	imgRef, err := CalculateTemporaryImageTag(bunRef)
	require.NoError(t, err)
	assert.Equal(t, "example.com/mybuns:porter-a3ffa80326b2e3a64e9ed3f377204ab2", imgRef.String())
}
