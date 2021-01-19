package extensions

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
)

func TestSupportsExtension(t *testing.T) {
	t.Run("key present", func(t *testing.T) {
		b := bundle.Bundle{RequiredExtensions: []string{"io.test.thing"}}
		assert.True(t, SupportsExtension(b, "io.test.thing"))
	})

	t.Run("key missing", func(t *testing.T) {
		// We need to match against the full key, not just shorthand
		b := bundle.Bundle{RequiredExtensions: []string{"thing"}}
		assert.False(t, SupportsExtension(b, "io.test.thing"))
	})
}
