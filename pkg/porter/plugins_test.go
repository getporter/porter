package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/plugins"
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
	t.Run("table", func(t *testing.T) {
		p := NewTestPorter(t)
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

	t.Run("yaml", func(t *testing.T) {
		p := NewTestPorter(t)
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

	t.Run("json", func(t *testing.T) {
		p := NewTestPorter(t)
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

func TestPorter_ShowPlugin(t *testing.T) {
	t.Run("table", func(t *testing.T) {
		p := NewTestPorter(t)
		opts := ShowPluginOptions{Name: "plugin1"}
		opts.Format = printer.FormatTable
		err := p.ShowPlugin(opts)
		require.NoError(t, err, "ShowPlugin failed")

		expected := `Name: plugin1
Version: v1.0
Commit: abc123
Author: Porter Authors

---------------------------
  Type     Implementation  
---------------------------
  storage  blob            
  storage  mongo           
`
		actual := p.TestConfig.TestContext.GetOutput()
		assert.Equal(t, expected, actual)
	})

	t.Run("yaml", func(t *testing.T) {
		p := NewTestPorter(t)
		opts := ShowPluginOptions{Name: "plugin1"}
		opts.Format = printer.FormatYaml
		err := p.ShowPlugin(opts)
		require.NoError(t, err, "ShowPlugin failed")

		expected := `name: plugin1
versioninfo:
  version: v1.0
  commit: abc123
  author: Porter Authors
implementations:
- type: storage
  name: blob
- type: storage
  name: mongo

`
		actual := p.TestConfig.TestContext.GetOutput()
		assert.Equal(t, expected, actual)
	})

	t.Run("json", func(t *testing.T) {
		p := NewTestPorter(t)
		opts := ShowPluginOptions{Name: "plugin1"}
		opts.Format = printer.FormatJson
		err := p.ShowPlugin(opts)
		require.NoError(t, err, "ShowPlugin failed")

		expected := `{
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
}
`
		actual := p.TestConfig.TestContext.GetOutput()
		assert.Equal(t, expected, actual)
	})
}

func TestPorter_InstallPlugin(t *testing.T) {
	p := NewTestPorter(t)

	opts := plugins.InstallOptions{}
	opts.URL = "https://example.com"
	err := opts.Validate([]string{"plugin1"})
	require.NoError(t, err, "Validate failed")

	err = p.InstallPlugin(opts)
	require.NoError(t, err, "InstallPlugin failed")

	wantOutput := "installed plugin1 plugin v1.0 (abc123)"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotOutput)
}

func TestPorter_UninstallPlugin(t *testing.T) {
	p := NewTestPorter(t)

	opts := pkgmgmt.UninstallOptions{}
	err := opts.Validate([]string{"plugin1"})
	require.NoError(t, err, "Validate failed")

	err = p.UninstallPlugin(opts)
	require.NoError(t, err, "UninstallPlugin failed")

	wantOutput := "Uninstalled plugin1 plugin"
	gotoutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotoutput)
}
