package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallOptions_validateParams(t *testing.T) {
	opts := InstallOptions{
		Params: []string{"A=1", "B=2"},
	}

	err := opts.validateParams()
	require.NoError(t, err)

	assert.Len(t, opts.Params, 2)
}

func TestInstallOptions_validateClaimName(t *testing.T) {
	testcases := []struct {
		name      string
		args      []string
		wantClaim string
		wantError string
	}{
		{"none", nil, "", ""},
		{"name set", []string{"wordpress"}, "wordpress", ""},
		{"too many args", []string{"wordpress", "extra"}, "", "only one positional argument may be specified, the claim name, but multiple were received: [wordpress extra]"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := InstallOptions{}
			err := opts.validateClaimName(tc.args)

			if tc.wantError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantClaim, opts.Name)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
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
