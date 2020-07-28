package extensions

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadDependencyProperties(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/bundle.json")
	require.NoError(t, err, "cannot read bundle file")

	bun, err := bundle.Unmarshal(data)
	require.NoError(t, err, "could not unmarshal the bundle")

	deps, err := ReadDependencies(bun)

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

func TestSortDependenciesBySequence(t *testing.T) {
	sequenceMock := []string{"nginx", "storage", "mysql"}

	bun := &bundle.Bundle{
		Custom: map[string]interface{}{
			DependenciesKey: Dependencies{
				Sequence: sequenceMock,
				Requires: map[string]Dependency{
					"mysql": Dependency{
						Bundle: "somecloud/mysql",
						Version: &DependencyVersion{
							AllowPrereleases: true,
							Ranges:           []string{"5.7.x"},
						},
					},
					"storage": Dependency{
						Bundle: "somecloud/blob-storage",
					},
					"nginx": {
						Bundle: "localhost:5000/nginx:1.19",
					},
				},
			},
		},
	}

	deps, err := ReadDependencies(bun)
	require.NoError(t, err, "unable to read dependencies extension data")

	assert.NotNil(t, deps, "Dependencies was not populated")
	assert.Len(t, deps.Requires, 3, "Dependencies.Requires is the wrong length")

	dep := deps.Requires["storage"]
	assert.NotNil(t, dep, "expected Dependencies.Requires to have an entry for 'storage'")

	dep = deps.Requires["mysql"]
	assert.NotNil(t, dep, "expected Dependencies.Requires to have an entry for 'mysql'")

	dep = deps.Requires["nginx"]
	assert.NotNil(t, dep, "expected Dependencies.Requires to have an entry for 'nginx'")

	// Get the keys out of the deps.Requires map
	keys := reflect.ValueOf(deps.Requires).MapKeys()

	// assert the bundles are sorted as sequenced
	assert.EqualValues(t, sequenceMock[0], keys[0].Interface().(string))
	assert.EqualValues(t, sequenceMock[1], keys[1].Interface().(string))
	assert.EqualValues(t, sequenceMock[2], keys[2].Interface().(string))
}
