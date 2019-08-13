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
	assert.Equal(t, "mysql", lock.Name)
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
			dep:       Dependency{"mysql", &DependencyVersion{nil, true}},
			wantError: "not implemented"},
		{name: "unimplemented missing tag",
			dep:       Dependency{Bundle: "mysql"},
			wantError: "not implemented"},
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
