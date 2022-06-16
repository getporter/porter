package cnab

import (
	"io/ioutil"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadDependencyProperties(t *testing.T) {
	t.Parallel()

	data, err := ioutil.ReadFile("testdata/bundle.json")
	require.NoError(t, err, "cannot read bundle file")

	b, err := bundle.Unmarshal(data)
	require.NoError(t, err, "could not unmarshal the bundle")

	bun := NewBundle(*b)
	assert.True(t, bun.HasDependencies())

	deps, err := bun.ReadDependencies()

	assert.NotNil(t, deps, "Dependencies was not populated")
	assert.Len(t, deps.Requires, 2, "Dependencies.Requires is the wrong length")

	dep := deps.Requires["storage"]
	assert.NotNil(t, dep, "expected Dependencies.Requires to have an entry for 'storage'")
	assert.Equal(t, "somecloud/blob-storage", dep.Bundle, "Dependency.Bundle is incorrect")
	assert.Nil(t, dep.Version, "Dependency.Version should be nil")

	dep = deps.Requires["mysql"]
	assert.NotNil(t, dep, "expected Dependencies.Requires to have an entry for 'mysql'")
	assert.Equal(t, "somecloud/mysql", dep.Bundle, "Dependency.Bundle is incorrect")
	assert.True(t, dep.Version.AllowPrereleases, "Dependency.Bundle.Version.AllowPrereleases should be true")
	assert.Equal(t, []string{"5.7.x"}, dep.Version.Ranges, "Dependency.Bundle.Version.Ranges is incorrect")

}

func TestDependencies_ListBySequence(t *testing.T) {
	t.Parallel()

	sequenceMock := []string{"nginx", "storage", "mysql"}

	bun := NewBundle(bundle.Bundle{
		Custom: map[string]interface{}{
			DependenciesExtensionKey: Dependencies{
				Sequence: sequenceMock,
				Requires: map[string]Dependency{
					"mysql": Dependency{
						Name:   "mysql",
						Bundle: "somecloud/mysql",
						Version: &DependencyVersion{
							AllowPrereleases: true,
							Ranges:           []string{"5.7.x"},
						},
					},
					"storage": Dependency{
						Name:   "storage",
						Bundle: "somecloud/blob-storage",
					},
					"nginx": Dependency{
						Name:   "nginx",
						Bundle: "localhost:5000/nginx:1.19",
					},
				},
			},
		},
	})

	rawDeps, err := bun.ReadDependencies()
	orderedDeps := rawDeps.ListBySequence()

	require.NoError(t, err, "unable to read dependencies extension data")

	assert.NotNil(t, orderedDeps, "Dependencies was not populated")
	assert.Len(t, orderedDeps, 3, "Dependencies.Requires is the wrong length")

	assert.NotNil(t, orderedDeps[0], "expected Dependencies.Requires to have an entry for 'storage")
	assert.NotNil(t, orderedDeps[1], "expected Dependencies.Requires to have an entry for 'mysql'")
	assert.NotNil(t, orderedDeps[2], "expected Dependencies.Requires to have an entry for 'nginx'")

	// assert the bundles are sorted as sequenced
	assert.Equal(t, sequenceMock[0], orderedDeps[0].Name, "unexpected order of the dependencies")
	assert.Equal(t, sequenceMock[1], orderedDeps[1].Name, "unexpected order of the dependencies")
	assert.Equal(t, sequenceMock[2], orderedDeps[2].Name, "unexpected order of the dependencies")
}

func TestSupportsDependencies(t *testing.T) {
	t.Parallel()

	t.Run("supported", func(t *testing.T) {
		b := NewBundle(bundle.Bundle{
			RequiredExtensions: []string{DependenciesExtensionKey},
		})

		assert.True(t, b.SupportsDependencies())
	})
	t.Run("unsupported", func(t *testing.T) {
		b := ExtendedBundle{}

		assert.False(t, b.SupportsDependencies())
	})
}

func TestHasDependencies(t *testing.T) {
	t.Parallel()

	t.Run("has dependencies", func(t *testing.T) {
		b := NewBundle(bundle.Bundle{
			RequiredExtensions: []string{DependenciesExtensionKey},
			Custom: map[string]interface{}{
				DependenciesExtensionKey: struct{}{},
			},
		})

		assert.True(t, b.HasDependencies())
	})
	t.Run("no dependencies", func(t *testing.T) {
		b := NewBundle(bundle.Bundle{
			RequiredExtensions: []string{DependenciesExtensionKey},
		})

		assert.False(t, b.HasDependencies())
	})
}
