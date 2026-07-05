package experimental

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags_FileSources(t *testing.T) {
	t.Run("sets FlagFileSources when file-sources is present", func(t *testing.T) {
		flags := ParseFlags([]string{FileSources})
		assert.Equal(t, FlagFileSources, flags&FlagFileSources)
	})

	t.Run("does not set FlagFileSources when absent", func(t *testing.T) {
		flags := ParseFlags([]string{NoopFeature})
		assert.Equal(t, FeatureFlags(0), flags&FlagFileSources)
	})

	t.Run("sets multiple flags independently", func(t *testing.T) {
		flags := ParseFlags([]string{PersistentParameters, FileSources})
		assert.Equal(t, FlagPersistentParameters, flags&FlagPersistentParameters)
		assert.Equal(t, FlagFileSources, flags&FlagFileSources)
	})
}

func TestIsFeatureEnabled_FileSources(t *testing.T) {
	t.Run("returns true when FlagFileSources is set", func(t *testing.T) {
		flags := ParseFlags([]string{FileSources})
		assert.True(t, flags&FlagFileSources == FlagFileSources)
	})

	t.Run("returns false when FlagFileSources is not set", func(t *testing.T) {
		flags := ParseFlags([]string{})
		assert.False(t, flags&FlagFileSources == FlagFileSources)
	})
}
