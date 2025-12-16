package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInspectableDependency_Structure(t *testing.T) {
	dep := InspectableDependency{
		Alias:        "nginx",
		Reference:    "localhost:5000/nginx:v1.19",
		Version:      ">=1.19",
		Depth:        0,
		SharingMode:  true,
		SharingGroup: "web",
		Parameters: map[string]string{
			"port": "bundle.parameters.nginx-port",
		},
		Credentials: map[string]string{
			"token": "bundle.credentials.registry-token",
		},
		Outputs: map[string]string{
			"endpoint": "nginx-endpoint",
		},
		Dependencies: []InspectableDependency{
			{
				Alias:     "certbot",
				Reference: "getporter/certbot:v1.0",
				Version:   "^1.0",
				Depth:     1,
			},
		},
	}

	assert.Equal(t, "nginx", dep.Alias)
	assert.Equal(t, "localhost:5000/nginx:v1.19", dep.Reference)
	assert.Equal(t, ">=1.19", dep.Version)
	assert.Equal(t, 0, dep.Depth)
	assert.True(t, dep.SharingMode)
	assert.Equal(t, "web", dep.SharingGroup)
	assert.Len(t, dep.Parameters, 1)
	assert.Len(t, dep.Credentials, 1)
	assert.Len(t, dep.Outputs, 1)
	assert.Len(t, dep.Dependencies, 1)
	assert.Equal(t, "certbot", dep.Dependencies[0].Alias)
	assert.Equal(t, 1, dep.Dependencies[0].Depth)
}

func TestInspectableBundle_DependenciesField(t *testing.T) {
	ib := InspectableBundle{
		Name:        "my-bundle",
		Description: "Test bundle",
		Version:     "1.0.0",
		Dependencies: []InspectableDependency{
			{
				Alias:     "mysql",
				Reference: "getporter/mysql:v0.1.3",
				Version:   "~0.1.0",
				Depth:     0,
			},
		},
	}

	assert.Equal(t, "my-bundle", ib.Name)
	assert.Len(t, ib.Dependencies, 1)
	assert.Equal(t, "mysql", ib.Dependencies[0].Alias)
}

func TestExplainOpts_DependencyFields(t *testing.T) {
	opts := ExplainOpts{
		ShowDependencies:   true,
		MaxDependencyDepth: 10,
	}

	assert.True(t, opts.ShowDependencies)
	assert.Equal(t, 10, opts.MaxDependencyDepth)
}
