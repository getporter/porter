package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallOptions_Prepare(t *testing.T) {
	opts := InstallOptions{
		Params: []string{"A=1", "B=2"},
	}

	err := opts.Validate()
	require.NoError(t, err)

	assert.Len(t, opts.Params, 2)
}

func TestInstallOptions_combineParameters(t *testing.T) {
	opts := InstallOptions{
		ParamFiles: []string{
			"testdata/install/base-params.txt",
			"testdata/install/dev-params.txt",
		},
		Params: []string{"A=true", "E=puppies", "E=kitties"},
	}

	err := opts.validateParams()
	require.NoError(t, err)

	gotParams := opts.combineParameters()

	wantParams := map[string]string{
		"A": "true",
		"B": "2",
		"C": "3",
		"D": "blue",
		"E": "kitties",
	}

	assert.Equal(t, wantParams, gotParams)
}
