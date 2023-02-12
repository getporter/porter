package cnab

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadDependencyV2Properties(t *testing.T) {
	t.Parallel()

	bun := ReadTestBundle(t, "testdata/bundle-depsv2.json")
	require.True(t, bun.HasDependenciesV2())

	deps, err := bun.ReadDependenciesV2()
	require.NoError(t, err)

	require.NotNil(t, deps, "DependenciesV2 was not populated")
	assert.Len(t, deps.Requires, 2, "DependenciesV2.Requires is the wrong length")

	dep := deps.Requires["storage"]
	require.NotNil(t, dep, "expected DependenciesV2.Requires to have an entry for 'storage'")
	assert.Equal(t, "somecloud/blob-storage", dep.Bundle, "DependencyV2.Bundle is incorrect")
	assert.Empty(t, dep.Version, "DependencyV2.Version should be nil")

	dep = deps.Requires["mysql"]
	require.NotNil(t, dep, "expected DependenciesV2.Requires to have an entry for 'mysql'")
	assert.Equal(t, "somecloud/mysql", dep.Bundle, "DependencyV2.Bundle is incorrect")
	assert.Equal(t, "5.7.x", dep.Version, "DependencyV2.Bundle.Version is incorrect")

}

func TestSupportsDependenciesV2(t *testing.T) {
	t.Parallel()

	t.Run("supported", func(t *testing.T) {
		b := ExtendedBundle{bundle.Bundle{
			RequiredExtensions: []string{DependenciesV2ExtensionKey},
		}}

		assert.True(t, b.SupportsDependenciesV2())
	})
	t.Run("unsupported", func(t *testing.T) {
		b := ExtendedBundle{}

		assert.False(t, b.SupportsDependenciesV2())
	})
}

func TestHasDependenciesV2(t *testing.T) {
	t.Parallel()

	t.Run("has dependencies", func(t *testing.T) {
		b := ExtendedBundle{bundle.Bundle{
			RequiredExtensions: []string{DependenciesV2ExtensionKey},
			Custom: map[string]interface{}{
				DependenciesV2ExtensionKey: struct{}{},
			},
		}}

		assert.True(t, b.HasDependenciesV2())
	})
	t.Run("no dependencies", func(t *testing.T) {
		b := ExtendedBundle{bundle.Bundle{
			RequiredExtensions: []string{DependenciesV2ExtensionKey},
		}}

		assert.False(t, b.HasDependenciesV2())
	})
}
