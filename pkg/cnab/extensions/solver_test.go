package extensions

import (
	"testing"

	"github.com/deislabs/cnab-go/bundle"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencySolver_ResolveDependencies(t *testing.T) {

	bun := &bundle.Bundle{
		Custom: map[string]interface{}{
			DependenciesKey: Dependencies{
				Requires: map[string]Dependency{
					"mysql": {
						Bundle: "deislabs/mysql:5.7",
					},
				},
			},
		},
	}

	s := DependencySolver{}
	locks, err := s.ResolveDependencies(bun)
	require.NoError(t, err)

	require.Len(t, locks, 1)

	lock := locks[0]
	assert.Equal(t, "mysql", lock.Alias)
	assert.Equal(t, "deislabs/mysql:5.7", lock.Tag)
}

func TestDependencySolver_ResolveVersion(t *testing.T) {
	testcases := []struct {
		name        string
		dep         Dependency
		wantVersion string
		wantError   string
	}{
		{name: "pinned version",
			dep:         Dependency{"mysql:5.7", nil},
			wantVersion: "5.7"},
		{name: "unimplemented range",
			dep:       Dependency{"mysql", &DependencyVersion{Ranges: []string{"1 - 1.5"}}},
			wantError: "not implemented"},
		{name: "default tag to latest",
			dep:         Dependency{Bundle: "deislabs/porter-test-only-latest"},
			wantVersion: "latest"},
		{name: "no default tag",
			dep:       Dependency{Bundle: "deislabs/porter-test-no-default-tag"},
			wantError: "no tag was specified"},
		{name: "default tag to highest semver",
			dep:         Dependency{"deislabs/porter-test-with-versions", &DependencyVersion{Ranges: nil, AllowPrereleases: true}},
			wantVersion: "v1.3-beta1"},
		{name: "default tag to highest semver, explicitly excluding prereleases",
			dep:         Dependency{"deislabs/porter-test-with-versions", &DependencyVersion{Ranges: nil, AllowPrereleases: false}},
			wantVersion: "v1.2"},
		{name: "default tag to highest semver, excluding prereleases by default",
			dep:         Dependency{Bundle: "deislabs/porter-test-with-versions"},
			wantVersion: "v1.2"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := DependencySolver{}
			version, err := s.ResolveVersion("mysql", tc.dep)

			if tc.wantError != "" {
				require.Error(t, err, "ResolveVersion should have returned an error")
				assert.Contains(t, err.Error(), tc.wantError)
			} else {
				require.NoError(t, err, "ResolveVersion should not have returned an error")

				assert.Equal(t, tc.wantVersion, version, "incorrect version resolved")
			}
		})
	}
}
