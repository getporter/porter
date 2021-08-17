package cnab

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
)

func TestSupportsExtension(t *testing.T) {
	t.Run("key present", func(t *testing.T) {
		b := ExtendedBundle{bundle.Bundle{RequiredExtensions: []string{"io.test.thing"}}}
		assert.True(t, b.SupportsExtension("io.test.thing"))
	})

	t.Run("key missing", func(t *testing.T) {
		// We need to match against the full key, not just shorthand
		b := ExtendedBundle{bundle.Bundle{RequiredExtensions: []string{"thing"}}}
		assert.False(t, b.SupportsExtension("io.test.thing"))
	})
}
