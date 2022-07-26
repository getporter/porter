package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_PrintPlugins(t *testing.T) {
	t.Run("plaintext", func(t *testing.T) {
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		opts := PrintPluginsOptions{
			PrintOptions: printer.PrintOptions{
				Format: printer.FormatPlaintext,
			},
		}
		err := p.PrintPlugins(ctx, opts)

		require.Nil(t, err)

		got := p.TestConfig.TestContext.GetOutput()
		test.CompareGoldenFile(t, "testdata/plugins/list-output.txt", got)
	})

	t.Run("yaml", func(t *testing.T) {
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		opts := PrintPluginsOptions{
			PrintOptions: printer.PrintOptions{
				Format: printer.FormatYaml,
			},
		}
		err := p.PrintPlugins(ctx, opts)

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
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		opts := PrintPluginsOptions{
			PrintOptions: printer.PrintOptions{
				Format: printer.FormatJson,
			},
		}
		err := p.PrintPlugins(ctx, opts)

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
	t.Run("plaintext", func(t *testing.T) {
		ctx := context.Background()

		p := NewTestPorter(t)
		defer p.Close()

		opts := ShowPluginOptions{Name: "plugin1"}
		opts.Format = printer.FormatPlaintext
		err := p.ShowPlugin(ctx, opts)
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
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		opts := ShowPluginOptions{Name: "plugin1"}
		opts.Format = printer.FormatYaml
		err := p.ShowPlugin(ctx, opts)
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
		ctx := context.Background()

		p := NewTestPorter(t)
		defer p.Close()

		opts := ShowPluginOptions{Name: "plugin1"}
		opts.Format = printer.FormatJson
		err := p.ShowPlugin(ctx, opts)
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
	defer p.Close()

	opts := plugins.InstallOptions{}
	opts.URL = "https://example.com"
	err := opts.Validate([]string{"plugin1"})
	require.NoError(t, err, "Validate failed")

	err = p.InstallPlugin(context.Background(), opts)
	require.NoError(t, err, "InstallPlugin failed")

	wantOutput := "installed plugin1 plugin v1.0 (abc123)\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotOutput)
}

func TestPorter_UninstallPlugin(t *testing.T) {
	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	opts := pkgmgmt.UninstallOptions{}
	err := opts.Validate([]string{"plugin1"})
	require.NoError(t, err, "Validate failed")

	err = p.UninstallPlugin(ctx, opts)
	require.NoError(t, err, "UninstallPlugin failed")

	wantOutput := "Uninstalled plugin1 plugin"
	gotoutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotoutput)
}
