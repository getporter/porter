package extensions

import (
	"io/ioutil"
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
