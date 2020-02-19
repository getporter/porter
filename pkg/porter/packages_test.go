package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchOptions_Validate_PackageName(t *testing.T) {
	opts := SearchOptions{}

	err := opts.validatePackageName([]string{})
	require.NoError(t, err)
	assert.Equal(t, "", opts.Name)

	err = opts.validatePackageName([]string{"helm"})
	require.NoError(t, err)
	assert.Equal(t, "helm", opts.Name)

	err = opts.validatePackageName([]string{"helm", "nstuff"})
	require.EqualError(t, err, "only one positional argument may be specified, the package name, but multiple were received: [helm nstuff]")
}
