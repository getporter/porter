package pkgmgmt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUninstallOptions_Validate(t *testing.T) {
	t.Run("no name", func(t *testing.T) {
		opts := UninstallOptions{}
		err := opts.Validate(nil)
		require.EqualError(t, err, "no name was specified")
	})
	t.Run("name specified", func(t *testing.T) {
		opts := UninstallOptions{}
		err := opts.Validate([]string{"thename"})
		require.NoError(t, err)
		assert.Equal(t, "thename", opts.Name, "the package name was not captured")
	})
	t.Run("multiple names specified", func(t *testing.T) {
		opts := UninstallOptions{}
		err := opts.Validate([]string{"name1", "name2"})
		require.EqualError(t, err, "only one positional argument may be specified, the name, but multiple were received: [name1 name2]")
	})
}
