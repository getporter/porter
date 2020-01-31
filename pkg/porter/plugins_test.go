package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/crudstore"
	"get.porter.sh/porter/pkg/storage/filesystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInternalPluginOpts_Validate(t *testing.T) {
	cfg := config.NewTestConfig(t)
	var opts RunInternalPluginOpts

	t.Run("no key", func(t *testing.T) {
		err := opts.Validate(nil, cfg.Config)
		require.Error(t, err)
		assert.Equal(t, err.Error(), "The positional argument KEY was not specified")
	})

	t.Run("too many keys", func(t *testing.T) {
		err := opts.Validate([]string{"foo", "bar"}, cfg.Config)
		require.Error(t, err)
		assert.Equal(t, err.Error(), "Multiple positional arguments were specified but only one, KEY is expected")
	})

	t.Run("valid key", func(t *testing.T) {
		err := opts.Validate([]string{filesystem.PluginKey}, cfg.Config)
		require.NoError(t, err)
		assert.Equal(t, opts.selectedInterface, crudstore.PluginInterface)
		assert.NotNil(t, opts.selectedPlugin)
	})

	t.Run("invalid key", func(t *testing.T) {
		err := opts.Validate([]string{"foo"}, cfg.Config)
		require.Error(t, err)
		assert.Equal(t, err.Error(), `invalid plugin key specified: "foo"`)
	})
}
