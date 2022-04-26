package cnab

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencySolver_ResolveDependencies(t *testing.T) {
	t.Parallel()

	bun := ExtendedBundle{bundle.Bundle{
		Custom: map[string]interface{}{
			DependenciesExtensionKey: Dependencies{
				Requires: map[string]Dependency{
					"mysql": {
						Bundle: "getporter/mysql:5.7",
					},
					"nginx": {
						Bundle: "localhost:5000/nginx:1.19",
					},
				},
			},
		},
	}}

	s := DependencySolver{}
	locks, err := s.ResolveDependencies(bun)
	require.NoError(t, err)
	require.Len(t, locks, 2)

	var mysql DependencyLock
	var nginx DependencyLock
	for _, lock := range locks {
		switch lock.Alias {
		case "mysql":
			mysql = lock
		case "nginx":
			nginx = lock
		}
	}

	assert.Equal(t, "getporter/mysql:5.7", mysql.Reference)
	assert.Equal(t, "localhost:5000/nginx:1.19", nginx.Reference)
}

func TestDependencySolver_ResolveVersion(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name        string
		dep         Dependency
		wantVersion string
		wantError   string
	}{
		{name: "pinned version",
			dep:         Dependency{Bundle: "mysql:5.7"},
			wantVersion: "5.7"},
		{name: "unimplemented range",
			dep:       Dependency{Bundle: "mysql", Version: &DependencyVersion{Ranges: []string{"1 - 1.5"}}},
			wantError: "not implemented"},
		{name: "default tag to latest",
			dep:         Dependency{Bundle: "getporterci/porter-test-only-latest"},
			wantVersion: "latest"},
		{name: "no default tag",
			dep:       Dependency{Bundle: "getporterci/porter-test-no-default-tag"},
			wantError: "no tag was specified"},
		{name: "default tag to highest semver",
			dep:         Dependency{Bundle: "getporterci/porter-test-with-versions", Version: &DependencyVersion{Ranges: nil, AllowPrereleases: true}},
			wantVersion: "v1.3-beta1"},
		{name: "default tag to highest semver, explicitly excluding prereleases",
			dep:         Dependency{Bundle: "getporterci/porter-test-with-versions", Version: &DependencyVersion{Ranges: nil, AllowPrereleases: false}},
			wantVersion: "v1.2"},
		{name: "default tag to highest semver, excluding prereleases by default",
			dep:         Dependency{Bundle: "getporterci/porter-test-with-versions"},
			wantVersion: "v1.2"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			s := DependencySolver{}
			version, err := s.ResolveVersion("mysql", tc.dep)

			if tc.wantError != "" {
				require.Error(t, err, "ResolveVersion should have returned an error")
				assert.Contains(t, err.Error(), tc.wantError)
			} else {
				require.NoError(t, err, "ResolveVersion should not have returned an error")

				assert.Equal(t, tc.wantVersion, version.Tag(), "incorrect version resolved")
			}
		})
	}
}
