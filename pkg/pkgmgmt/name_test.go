package pkgmgmt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateName(t *testing.T) {
	t.Run("valid names", func(t *testing.T) {
		for _, name := range []string{"exec", "az-cli", "my_mixin", "mixin2"} {
			require.NoError(t, ValidateName(name), name)
		}
	})

	t.Run("invalid names", func(t *testing.T) {
		for _, name := range []string{"MyMixin", "my mixin", "my.mixin", ""} {
			err := ValidateName(name)
			require.Error(t, err, name)
			require.Contains(t, err.Error(), "must contain only lowercase letters, numbers, dashes, and underscores")
		}
	})
}
