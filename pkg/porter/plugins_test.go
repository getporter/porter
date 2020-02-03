package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/printer"
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

func TestPorter_PrintPlugins(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	t.Run("Plugin List - Table Format", func(t *testing.T) {

		opts := PrintPluginsOptions{
			PrintOptions: printer.PrintOptions{
				Format: printer.FormatTable,
			},
		}
		err := p.PrintPlugins(opts)

		require.Nil(t, err)
		expected := `Name      Version   Author
plugin1   v1.0      Porter Authors
plugin2   v1.0      Porter Authors
unknown   v1.0      Porter Authors
`
		actual := p.TestConfig.TestContext.GetOutput()
		assert.Equal(t, expected, actual)
	})

	t.Run("Plugin List - YAML Format", func(t *testing.T) {
		p.TestConfig.TestContext.ResetOutput()
		opts := PrintPluginsOptions{
			PrintOptions: printer.PrintOptions{
				Format: printer.FormatYaml,
			},
		}
		err := p.PrintPlugins(opts)

		require.Nil(t, err)
		expected := `- name: plugin1
  versioninfo:
    version: v1.0
    commit: abc123
    author: Porter Authors
  implementations:
  - type: storage
    name: blob
  - type: storage
    name: mongo
- name: plugin2
  versioninfo:
    version: v1.0
    commit: abc123
    author: Porter Authors
  implementations:
  - type: storage
    name: blob
  - type: storage
    name: mongo
- name: unknown
  versioninfo:
    version: v1.0
    commit: abc123
    author: Porter Authors
  implementations: []

`
		actual := p.TestConfig.TestContext.GetOutput()
		assert.Equal(t, expected, actual)
	})

	t.Run("Plugin List - JSON Format", func(t *testing.T) {
		p.TestConfig.TestContext.ResetOutput()
		opts := PrintPluginsOptions{
			PrintOptions: printer.PrintOptions{
				Format: printer.FormatJson,
			},
		}
		err := p.PrintPlugins(opts)

		require.Nil(t, err)
		expected := `[
  {
    "name": "plugin1",
    "version": "v1.0",
    "commit": "abc123",
    "author": "Porter Authors",
    "implementations": [
      {
        "type": "storage",
        "implementation": "blob"
      },
      {
        "type": "storage",
        "implementation": "mongo"
      }
    ]
  },
  {
    "name": "plugin2",
    "version": "v1.0",
    "commit": "abc123",
    "author": "Porter Authors",
    "implementations": [
      {
        "type": "storage",
        "implementation": "blob"
      },
      {
        "type": "storage",
        "implementation": "mongo"
      }
    ]
  },
  {
    "name": "unknown",
    "version": "v1.0",
    "commit": "abc123",
    "author": "Porter Authors",
    "implementations": null
  }
]
`
		actual := p.TestConfig.TestContext.GetOutput()
		assert.Equal(t, expected, actual)
	})
}
