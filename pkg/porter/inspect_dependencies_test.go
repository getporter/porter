package porter

import (
	"context"
	"encoding/json"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	depsv1 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v1"
	depsv2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatVersionV1(t *testing.T) {
	tests := []struct {
		name     string
		version  *depsv1.DependencyVersion
		expected string
	}{
		{
			name:     "nil version",
			version:  nil,
			expected: "",
		},
		{
			name: "single range",
			version: &depsv1.DependencyVersion{
				Ranges: []string{">=1.0.0"},
			},
			expected: ">=1.0.0",
		},
		{
			name: "multiple ranges",
			version: &depsv1.DependencyVersion{
				Ranges: []string{">=1.0.0", "<2.0.0"},
			},
			expected: ">=1.0.0 || <2.0.0",
		},
		{
			name: "with prereleases",
			version: &depsv1.DependencyVersion{
				Ranges:           []string{">=1.0.0"},
				AllowPrereleases: true,
			},
			expected: ">=1.0.0 (including prereleases)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVersionV1(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenDependencyTree(t *testing.T) {
	deps := []InspectableDependency{
		{
			Alias:     "dep1",
			Reference: "localhost:5000/dep1:v1",
			Depth:     0,
			Dependencies: []InspectableDependency{
				{
					Alias:     "dep2",
					Reference: "localhost:5000/dep2:v1",
					Depth:     1,
				},
			},
		},
		{
			Alias:     "dep3",
			Reference: "localhost:5000/dep3:v1",
			Depth:     0,
		},
	}

	flattened := flattenDependencyTree(deps)

	assert.Len(t, flattened, 3)
	assert.Equal(t, "dep1", flattened[0].Alias)
	assert.Equal(t, 0, flattened[0].Depth)
	assert.Equal(t, "dep2", flattened[1].Alias)
	assert.Equal(t, 1, flattened[1].Depth)
	assert.Equal(t, "dep3", flattened[2].Alias)
	assert.Equal(t, 0, flattened[2].Depth)
}

func TestBuildDependencyTree_V1(t *testing.T) {
	// Create a bundle with v1 dependencies
	// Note: Using explicit tag in bundle reference to avoid version resolution
	bun := cnab.ExtendedBundle{
		Bundle: bundle.Bundle{
			Name:    "test-bundle",
			Version: "1.0.0",
			Custom: map[string]interface{}{
				cnab.DependenciesV1ExtensionKey: map[string]interface{}{
					"requires": map[string]interface{}{
						"mysql": map[string]interface{}{
							"bundle": "getporter/mysql:v0.1.3",
						},
					},
				},
			},
		},
	}

	builder := NewDependencyTreeBuilder(10)
	deps, err := builder.BuildDependencyTree(context.Background(), bun)

	require.NoError(t, err)
	require.Len(t, deps, 1)

	assert.Equal(t, "mysql", deps[0].Alias)
	assert.Equal(t, "getporter/mysql:v0.1.3", deps[0].Reference)
	assert.Equal(t, 0, deps[0].Depth)
}

func TestBuildDependencyTree_V2(t *testing.T) {
	// Create v2 dependencies
	// Note: Using explicit tag in bundle reference to avoid version resolution
	v2Deps := depsv2.Dependencies{
		Requires: map[string]depsv2.Dependency{
			"nginx": {
				Bundle: "localhost:5000/nginx:v1.19.0",
				Parameters: map[string]string{
					"port": "bundle.parameters.nginx-port",
				},
				Credentials: map[string]string{
					"token": "bundle.credentials.registry-token",
				},
				Outputs: map[string]string{
					"endpoint": "nginx-endpoint",
				},
				Sharing: depsv2.SharingCriteria{
					Mode: true,
					Group: depsv2.SharingGroup{
						Name: "web",
					},
				},
			},
		},
	}

	v2JSON, err := json.Marshal(v2Deps)
	require.NoError(t, err)

	var v2Map map[string]interface{}
	err = json.Unmarshal(v2JSON, &v2Map)
	require.NoError(t, err)

	// Create a bundle with v2 dependencies
	bun := cnab.ExtendedBundle{
		Bundle: bundle.Bundle{
			Name:    "test-bundle",
			Version: "1.0.0",
			Custom: map[string]interface{}{
				cnab.DependenciesV2ExtensionKey: v2Map,
			},
		},
	}

	builder := NewDependencyTreeBuilder(10)
	deps, err := builder.BuildDependencyTree(context.Background(), bun)

	require.NoError(t, err)
	require.Len(t, deps, 1)

	assert.Equal(t, "nginx", deps[0].Alias)
	assert.Equal(t, "localhost:5000/nginx:v1.19.0", deps[0].Reference)
	assert.Equal(t, 0, deps[0].Depth)
	assert.True(t, deps[0].SharingMode)
	assert.Equal(t, "web", deps[0].SharingGroup)
	assert.Equal(t, "bundle.parameters.nginx-port", deps[0].Parameters["port"])
	assert.Equal(t, "bundle.credentials.registry-token", deps[0].Credentials["token"])
	assert.Equal(t, "nginx-endpoint", deps[0].Outputs["endpoint"])
}

func TestBuildDependencyTree_NoDependencies(t *testing.T) {
	// Create a bundle without dependencies
	bun := cnab.ExtendedBundle{
		Bundle: bundle.Bundle{
			Name:    "test-bundle",
			Version: "1.0.0",
		},
	}

	builder := NewDependencyTreeBuilder(10)
	deps, err := builder.BuildDependencyTree(context.Background(), bun)

	require.NoError(t, err)
	assert.Nil(t, deps)
}
