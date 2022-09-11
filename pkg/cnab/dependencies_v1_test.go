package cnab

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadDependencyV1Properties(t *testing.T) {
	t.Parallel()

	bun := ReadTestBundle(t, "testdata/bundle.json")
	require.True(t, bun.HasDependenciesV1())

	deps, err := bun.ReadDependenciesV1()
	require.NoError(t, err, "ReadDependenciesV1 failed")
	assert.NotNil(t, deps, "Dependencies was not populated")
	assert.Len(t, deps.Requires, 2, "Dependencies.Requires is the wrong length")

	dep := deps.Requires["storage"]
	require.NotNil(t, dep, "expected Dependencies.Requires to have an entry for 'storage'")
	assert.Equal(t, "somecloud/blob-storage", dep.Bundle, "Dependency.Bundle is incorrect")
	assert.Nil(t, dep.Version, "Dependency.Version should be nil")

	dep = deps.Requires["mysql"]
	require.NotNil(t, dep, "expected Dependencies.Requires to have an entry for 'mysql'")
	assert.Equal(t, "somecloud/mysql", dep.Bundle, "Dependency.Bundle is incorrect")
	assert.True(t, dep.Version.AllowPrereleases, "Dependency.Bundle.Version.AllowPrereleases should be true")
	assert.Equal(t, []string{"5.7.x"}, dep.Version.Ranges, "Dependency.Bundle.Version.Ranges is incorrect")

}

func TestSupportsDependenciesV1(t *testing.T) {
	t.Parallel()

	t.Run("supported", func(t *testing.T) {
		b := ExtendedBundle{bundle.Bundle{
			RequiredExtensions: []string{DependenciesV1ExtensionKey},
		}}

		assert.True(t, b.SupportsDependenciesV1())
	})
	t.Run("unsupported", func(t *testing.T) {
		b := ExtendedBundle{}

		assert.False(t, b.SupportsDependenciesV1())
	})
}

func TestHasDependenciesV1(t *testing.T) {
	t.Parallel()

	t.Run("has dependencies", func(t *testing.T) {
		b := ExtendedBundle{bundle.Bundle{
			RequiredExtensions: []string{DependenciesV1ExtensionKey},
			Custom: map[string]interface{}{
				DependenciesV1ExtensionKey: struct{}{},
			},
		}}

		assert.True(t, b.HasDependenciesV1())
	})
	t.Run("no dependencies", func(t *testing.T) {
		b := ExtendedBundle{bundle.Bundle{
			RequiredExtensions: []string{DependenciesV1ExtensionKey},
		}}

		assert.False(t, b.HasDependenciesV1())
	})
}
