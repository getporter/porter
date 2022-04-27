package cnab

import (
	"io/ioutil"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessedExtensions_GetParameterSourcesExtension(t *testing.T) {
	t.Parallel()

	t.Run("extension present", func(t *testing.T) {
		t.Parallel()

		var ps ParameterSources
		ps.SetParameterFromOutput("tfstate", "tfstate")
		processed := ProcessedExtensions{
			ParameterSourcesExtensionKey: ps,
		}

		ext, required, err := processed.GetParameterSources()
		require.NoError(t, err, "GetParameterSources failed")
		assert.True(t, required, "parameter-sources should be a required extension")
		assert.Equal(t, ps, ext, "parameter-sources was not populated properly")
	})

	t.Run("extension missing", func(t *testing.T) {
		t.Parallel()

		processed := ProcessedExtensions{}

		ext, required, err := processed.GetParameterSources()
		require.NoError(t, err, "GetParameterSources failed")
		assert.False(t, required, "parameter-sources should NOT be a required extension")
		assert.Empty(t, ext, "parameter-sources should default to empty when not required")
	})

	t.Run("extension invalid", func(t *testing.T) {
		t.Parallel()

		processed := ProcessedExtensions{
			ParameterSourcesExtensionKey: map[string]string{"ponies": "are great"},
		}

		ext, required, err := processed.GetParameterSources()
		require.Error(t, err, "GetParameterSources should have failed")
		assert.True(t, required, "parameter-sources should be a required extension")
		assert.Empty(t, ext, "parameter-sources should default to empty when not required")
	})
}

func TestReadParameterSourcesProperties(t *testing.T) {
	t.Parallel()

	data, err := ioutil.ReadFile("testdata/bundle.json")
	require.NoError(t, err, "cannot read bundle file")

	b, err := bundle.Unmarshal(data)
	require.NoError(t, err, "could not unmarshal the bundle")
	bun := ExtendedBundle{*b}
	assert.True(t, bun.HasParameterSources())

	ps, err := bun.ReadParameterSources()
	require.NoError(t, err, "could not read parameter sources")

	want := ParameterSources{}
	want.SetParameterFromOutput("tfstate", "tfstate")
	want.SetParameterFromDependencyOutput("mysql_connstr", "mysql", "connstr")
	assert.Equal(t, want, ps)
}

func TestParameterSource_ListSourcesByPriority(t *testing.T) {
	t.Parallel()

	ps := ParameterSources{}
	ps.SetParameterFromOutput("tfstate", "tfstate")
	got := ps["tfstate"].ListSourcesByPriority()
	want := []ParameterSourceDefinition{
		OutputParameterSource{OutputName: "tfstate"},
	}
	assert.Equal(t, want, got)
}

func TestSupportsParameterSources(t *testing.T) {
	t.Parallel()

	t.Run("supported", func(t *testing.T) {
		b := ExtendedBundle{bundle.Bundle{
			RequiredExtensions: []string{ParameterSourcesExtensionKey},
		}}

		assert.True(t, b.SupportsParameterSources())
	})
	t.Run("unsupported", func(t *testing.T) {
		b := ExtendedBundle{}

		assert.False(t, b.SupportsParameterSources())
	})
}

func TestHasParameterSources(t *testing.T) {
	t.Parallel()

	t.Run("has parameter sources", func(t *testing.T) {
		b := ExtendedBundle{bundle.Bundle{
			RequiredExtensions: []string{ParameterSourcesExtensionKey},
			Custom: map[string]interface{}{
				ParameterSourcesExtensionKey: struct{}{},
			},
		}}

		assert.True(t, b.HasParameterSources())
	})
	t.Run("no parameter sources", func(t *testing.T) {
		b := ExtendedBundle{bundle.Bundle{
			RequiredExtensions: []string{ParameterSourcesExtensionKey},
		}}

		assert.False(t, b.HasParameterSources())
	})
}
