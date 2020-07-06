package extensions

import (
	"io/ioutil"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/require"
)

func TestProcessRequiredExtensions(t *testing.T) {
	t.Run("supported", func(t *testing.T) {
		data, err := ioutil.ReadFile("testdata/bundle.json")
		require.NoError(t, err, "cannot read bundle file")

		bun, err := bundle.Unmarshal(data)
		require.NoError(t, err, "could not unmarshal the bundle")

		exts, err := ProcessRequiredExtensions(*bun)
		require.NoError(t, err, "could not process required extensions")

		expected := ProcessedExtensions{
			"io.cnab.dependencies": &Dependencies{
				Requires: map[string]Dependency{
					"storage": Dependency{
						Bundle: "somecloud/blob-storage",
					},
					"mysql": Dependency{
						Bundle: "somecloud/mysql",
						Version: &DependencyVersion{
							AllowPrereleases: true,
							Ranges:           []string{"5.7.x"},
						},
					},
				},
			},
		}
		require.Equal(t, expected, exts)
	})

	t.Run("supported unprocessable", func(t *testing.T) {
		data, err := ioutil.ReadFile("testdata/bundle-supported-unprocessable.json")
		require.NoError(t, err, "cannot read bundle file")

		bun, err := bundle.Unmarshal(data)
		require.NoError(t, err, "could not unmarshal the bundle")

		_, err = ProcessRequiredExtensions(*bun)
		require.EqualError(t, err, "unable to process extension: io.cnab.docker: no custom extension configuration found")
	})

	t.Run("unsupported", func(t *testing.T) {
		data, err := ioutil.ReadFile("testdata/bundle-unsupported-required.json")
		require.NoError(t, err, "cannot read bundle file")

		bun, err := bundle.Unmarshal(data)
		require.NoError(t, err, "could not unmarshal the bundle")

		_, err = ProcessRequiredExtensions(*bun)
		require.EqualError(t, err, "unsupported required extension: donuts")
	})
}

func TestGetSupportedExtension(t *testing.T) {
	t.Run("supported via shorthand", func(t *testing.T) {
		ext, err := GetSupportedExtension("docker")
		require.NoError(t, err)
		require.Equal(t, DockerExtension.Key, ext.Key)
	})

	t.Run("supported via full key", func(t *testing.T) {
		ext, err := GetSupportedExtension("io.cnab.docker")
		require.NoError(t, err)
		require.Equal(t, DockerExtension.Key, ext.Key)
	})

	t.Run("unsupported", func(t *testing.T) {
		_, err := GetSupportedExtension("donuts")
		require.EqualError(t, err, "unsupported required extension: donuts")
	})
}
