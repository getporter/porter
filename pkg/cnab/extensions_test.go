package cnab

import (
	"fmt"
	"testing"

	depsv1ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v1"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessRequiredExtensions(t *testing.T) {
	t.Parallel()

	t.Run("supported", func(t *testing.T) {
		t.Parallel()

		bun := ReadTestBundle(t, "testdata/bundle.json")
		exts, err := bun.ProcessRequiredExtensions()
		require.NoError(t, err, "could not process required extensions")

		expected := ProcessedExtensions{
			"sh.porter.file-parameters": nil,
			"io.cnab.dependencies": depsv1ext.Dependencies{
				Requires: map[string]depsv1ext.Dependency{
					"storage": depsv1ext.Dependency{
						Bundle: "somecloud/blob-storage",
					},
					"mysql": depsv1ext.Dependency{
						Bundle: "somecloud/mysql",
						Version: &depsv1ext.DependencyVersion{
							AllowPrereleases: true,
							Ranges:           []string{"5.7.x"},
						},
					},
				},
			},
			"io.cnab.parameter-sources": ParameterSources{
				"tfstate": ParameterSource{
					Priority: []string{ParameterSourceTypeOutput},
					Sources: ParameterSourceMap{
						ParameterSourceTypeOutput: OutputParameterSource{"tfstate"},
					},
				},
				"mysql_connstr": ParameterSource{
					Priority: []string{ParameterSourceTypeDependencyOutput},
					Sources: ParameterSourceMap{
						ParameterSourceTypeDependencyOutput: DependencyOutputParameterSource{
							Dependency: "mysql",
							OutputName: "connstr",
						},
					},
				},
			},
		}
		require.Equal(t, expected, exts)
	})

	t.Run("supported unprocessable", func(t *testing.T) {
		t.Parallel()

		bun := ReadTestBundle(t, "testdata/bundle-supported-unprocessable.json")
		_, err := bun.ProcessRequiredExtensions()
		require.EqualError(t, err, "unable to process extension: io.cnab.docker: no custom extension configuration found")
	})

	t.Run("unsupported", func(t *testing.T) {
		t.Parallel()

		bun := ReadTestBundle(t, "testdata/bundle-unsupported-required.json")
		_, err := bun.ProcessRequiredExtensions()
		require.EqualError(t, err, "unsupported required extension: donuts")
	})
}

func TestGetSupportedExtension(t *testing.T) {
	t.Parallel()

	for _, supported := range SupportedExtensions {
		t.Run(fmt.Sprintf("%s - shorthand", supported.Shorthand), func(t *testing.T) {
			t.Parallel()

			ext, err := GetSupportedExtension(supported.Shorthand)
			require.NoError(t, err)
			require.Equal(t, supported.Key, ext.Key)
		})

		t.Run(fmt.Sprintf("%s - key", supported.Key), func(t *testing.T) {
			t.Parallel()

			ext, err := GetSupportedExtension(supported.Key)
			require.NoError(t, err)
			require.Equal(t, supported.Key, ext.Key)
		})
	}

	t.Run("unsupported", func(t *testing.T) {
		t.Parallel()

		_, err := GetSupportedExtension("donuts")
		require.EqualError(t, err, "unsupported required extension: donuts")
	})
}

func TestSupportsExtension(t *testing.T) {
	t.Run("key present", func(t *testing.T) {
		b := NewBundle(bundle.Bundle{RequiredExtensions: []string{"io.test.thing"}})
		assert.True(t, b.SupportsExtension("io.test.thing"))
	})

	t.Run("key missing", func(t *testing.T) {
		// We need to match against the full key, not just shorthand
		b := NewBundle(bundle.Bundle{RequiredExtensions: []string{"thing"}})
		assert.False(t, b.SupportsExtension("io.test.thing"))
	})
}
