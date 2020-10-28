package context

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_Debug(t *testing.T) {
	t.Run("porter_debug unset", func(t *testing.T) {
		os.Unsetenv("PORTER_DEBUG")

		c := New()
		assert.False(t, c.Debug, "debug should be false when the env var is not present")
	})

	t.Run("porter_debug false", func(t *testing.T) {
		defer os.Unsetenv("PORTER_DEBUG")
		os.Setenv("PORTER_DEBUG", "false")
		c := New()
		assert.False(t, c.Debug, "debug should be false when the env var is set to 'false'")
	})

	t.Run("porter_debug invalid", func(t *testing.T) {
		defer os.Unsetenv("PORTER_DEBUG")
		os.Setenv("PORTER_DEBUG", "blorp")
		c := New()
		assert.False(t, c.Debug, "debug should be false when the env var is not a bool")
	})

	t.Run("porter_debug true", func(t *testing.T) {
		os.Setenv("PORTER_DEBUG", "true")
		c := New()
		assert.True(t, c.Debug, "debug should be true when the env var is 'true'")
	})
}
