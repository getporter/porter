package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_applyDefaultOptions(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	err := p.Create()
	require.NoError(t, err)

	opts := &InstallOptions{}
	err = opts.validateParams()
	require.NoError(t, err)

	p.Debug = true
	err = p.applyDefaultOptions(&opts.sharedOptions)
	require.NoError(t, err)

	assert.Equal(t, p.Manifest.Name, opts.Name)

	debug, set := opts.combinedParameters["porter-debug"]
	assert.True(t, set)
	assert.Equal(t, "true", debug)
}

func TestPorter_applyDefaultOptions_NoManifest(t *testing.T) {
	p := NewTestPorter(t)

	opts := &InstallOptions{}
	err := opts.validateParams()
	require.NoError(t, err)

	err = p.applyDefaultOptions(&opts.sharedOptions)
	require.NoError(t, err)

	assert.Equal(t, "", opts.Name)
}

func TestPorter_applyDefaultOptions_DebugOff(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	err := p.Create()
	require.NoError(t, err)

	opts := InstallOptions{}
	err = opts.validateParams()
	require.NoError(t, err)

	p.Debug = false
	err = p.applyDefaultOptions(&opts.sharedOptions)
	require.NoError(t, err)

	assert.Equal(t, p.Manifest.Name, opts.Name)

	_, set := opts.combinedParameters["porter-debug"]
	assert.False(t, set)
}

func TestPorter_applyDefaultOptions_ParamSet(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	err := p.Create()
	require.NoError(t, err)

	opts := InstallOptions{
		sharedOptions{
			Params: []string{"porter-debug=false"},
		},
	}
	err = opts.validateParams()
	require.NoError(t, err)

	p.Debug = true
	err = p.applyDefaultOptions(&opts.sharedOptions)
	require.NoError(t, err)

	debug, set := opts.combinedParameters["porter-debug"]
	assert.True(t, set)
	assert.Equal(t, "false", debug)
}

func TestInstallOptions_validateParams(t *testing.T) {
	opts := InstallOptions{
		sharedOptions{
			Params: []string{"A=1", "B=2"},
		},
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
		sharedOptions{
			ParamFiles: []string{
				"testdata/install/base-params.txt",
				"testdata/install/dev-params.txt",
			},
			Params: []string{"A=true", "E=puppies", "E=kitties"},
		},
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
