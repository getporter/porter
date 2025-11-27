package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOutputReference(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected *OutputReference
	}{
		{
			name:  "valid output reference",
			value: "${bundle.dependencies.database.outputs.connstr}",
			expected: &OutputReference{
				DependencyName: "database",
				OutputName:     "connstr",
			},
		},
		{
			name:  "valid output reference with dashes",
			value: "${bundle.dependencies.my-database.outputs.connection-string}",
			expected: &OutputReference{
				DependencyName: "my-database",
				OutputName:     "connection-string",
			},
		},
		{
			name:  "valid output reference with underscores",
			value: "${bundle.dependencies.db_server.outputs.db_url}",
			expected: &OutputReference{
				DependencyName: "db_server",
				OutputName:     "db_url",
			},
		},
		{
			name:     "not an output reference - parameter",
			value:    "${bundle.parameters.foo}",
			expected: nil,
		},
		{
			name:     "not an output reference - credential",
			value:    "${bundle.credentials.bar}",
			expected: nil,
		},
		{
			name:     "not an output reference - literal value",
			value:    "some-literal-value",
			expected: nil,
		},
		{
			name:     "not an output reference - empty string",
			value:    "",
			expected: nil,
		},
		{
			name:     "not an output reference - malformed template",
			value:    "${bundle.dependencies.foo}",
			expected: nil,
		},
		{
			name:  "output reference with whitespace",
			value: "  ${bundle.dependencies.database.outputs.connstr}  ",
			expected: &OutputReference{
				DependencyName: "database",
				OutputName:     "connstr",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOutputReference(tt.value)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.DependencyName, result.DependencyName)
				assert.Equal(t, tt.expected.OutputName, result.OutputName)
			}
		})
	}
}

func TestBuildDependencyGraph_SimpleChain(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	m := &manifest.Manifest{
		Dependencies: manifest.Dependencies{
			Requires: []*manifest.Dependency{
				{
					Name: "database",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/mysql:v1.0.0",
					},
				},
				{
					Name: "app",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/webapp:v1.0.0",
					},
					Parameters: map[string]string{
						"db-connstr": "${bundle.dependencies.database.outputs.connstr}",
					},
				},
			},
		},
	}

	graph, err := p.Porter.buildDependencyGraph(m)
	require.NoError(t, err)
	require.NotNil(t, graph)

	// Verify nodes were created
	assert.Len(t, graph.Nodes, 2)
	assert.Contains(t, graph.Nodes, "database")
	assert.Contains(t, graph.Nodes, "app")

	// Verify app depends on database
	appNode := graph.Nodes["app"]
	assert.Len(t, appNode.Dependencies, 1)
	assert.Equal(t, "database", appNode.Dependencies[0].Name)

	// Verify database has app as dependent
	dbNode := graph.Nodes["database"]
	assert.Len(t, dbNode.Dependents, 1)
	assert.Equal(t, "app", dbNode.Dependents[0].Name)

	// Verify output reference is tracked
	assert.Len(t, appNode.OutputsUsed, 1)
	assert.Contains(t, appNode.OutputsUsed, "db-connstr")
	assert.Equal(t, "database", appNode.OutputsUsed["db-connstr"].DependencyName)
	assert.Equal(t, "connstr", appNode.OutputsUsed["db-connstr"].OutputName)
}

func TestBuildDependencyGraph_Diamond(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	m := &manifest.Manifest{
		Dependencies: manifest.Dependencies{
			Requires: []*manifest.Dependency{
				{
					Name: "infra",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/infra:v1.0.0",
					},
				},
				{
					Name: "database",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/mysql:v1.0.0",
					},
					Parameters: map[string]string{
						"vpc-id": "${bundle.dependencies.infra.outputs.vpc-id}",
					},
				},
				{
					Name: "cache",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/redis:v1.0.0",
					},
					Parameters: map[string]string{
						"vpc-id": "${bundle.dependencies.infra.outputs.vpc-id}",
					},
				},
				{
					Name: "app",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/webapp:v1.0.0",
					},
					Parameters: map[string]string{
						"db-connstr":    "${bundle.dependencies.database.outputs.connstr}",
						"cache-connstr": "${bundle.dependencies.cache.outputs.connstr}",
					},
				},
			},
		},
	}

	graph, err := p.Porter.buildDependencyGraph(m)
	require.NoError(t, err)
	require.NotNil(t, graph)

	// Verify nodes were created
	assert.Len(t, graph.Nodes, 4)

	// Verify infra has no dependencies
	infraNode := graph.Nodes["infra"]
	assert.Len(t, infraNode.Dependencies, 0)
	assert.Len(t, infraNode.Dependents, 2)

	// Verify database depends on infra
	dbNode := graph.Nodes["database"]
	assert.Len(t, dbNode.Dependencies, 1)
	assert.Equal(t, "infra", dbNode.Dependencies[0].Name)

	// Verify cache depends on infra
	cacheNode := graph.Nodes["cache"]
	assert.Len(t, cacheNode.Dependencies, 1)
	assert.Equal(t, "infra", cacheNode.Dependencies[0].Name)

	// Verify app depends on both database and cache
	appNode := graph.Nodes["app"]
	assert.Len(t, appNode.Dependencies, 2)
	depNames := []string{appNode.Dependencies[0].Name, appNode.Dependencies[1].Name}
	assert.Contains(t, depNames, "database")
	assert.Contains(t, depNames, "cache")
}

func TestBuildDependencyGraph_NoDependencies(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	m := &manifest.Manifest{
		Dependencies: manifest.Dependencies{
			Requires: []*manifest.Dependency{
				{
					Name: "database",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/mysql:v1.0.0",
					},
				},
				{
					Name: "cache",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/redis:v1.0.0",
					},
				},
			},
		},
	}

	graph, err := p.Porter.buildDependencyGraph(m)
	require.NoError(t, err)
	require.NotNil(t, graph)

	// Verify nodes were created
	assert.Len(t, graph.Nodes, 2)

	// Verify no dependencies between nodes
	for _, node := range graph.Nodes {
		assert.Len(t, node.Dependencies, 0)
		assert.Len(t, node.Dependents, 0)
		assert.Len(t, node.OutputsUsed, 0)
	}
}

func TestBuildDependencyGraph_MissingDependency(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	m := &manifest.Manifest{
		Dependencies: manifest.Dependencies{
			Requires: []*manifest.Dependency{
				{
					Name: "app",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/webapp:v1.0.0",
					},
					Parameters: map[string]string{
						"db-connstr": "${bundle.dependencies.database.outputs.connstr}",
					},
				},
			},
		},
	}

	graph, err := p.Porter.buildDependencyGraph(m)
	assert.Error(t, err)
	assert.Nil(t, graph)
	assert.Contains(t, err.Error(), "non-existent dependency database")
}

func TestBuildDependencyGraph_CredentialOutputs(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	m := &manifest.Manifest{
		Dependencies: manifest.Dependencies{
			Requires: []*manifest.Dependency{
				{
					Name: "database",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/mysql:v1.0.0",
					},
				},
				{
					Name: "app",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/webapp:v1.0.0",
					},
					Credentials: map[string]string{
						"db-password": "${bundle.dependencies.database.outputs.admin-password}",
					},
				},
			},
		},
	}

	graph, err := p.Porter.buildDependencyGraph(m)
	require.NoError(t, err)
	require.NotNil(t, graph)

	// Verify app depends on database
	appNode := graph.Nodes["app"]
	assert.Len(t, appNode.Dependencies, 1)
	assert.Equal(t, "database", appNode.Dependencies[0].Name)

	// Verify credential output reference is tracked
	assert.Len(t, appNode.OutputsUsed, 1)
	assert.Contains(t, appNode.OutputsUsed, "db-password")
	assert.Equal(t, "database", appNode.OutputsUsed["db-password"].DependencyName)
	assert.Equal(t, "admin-password", appNode.OutputsUsed["db-password"].OutputName)
}

func TestComputeExecutionOrder_SimpleChain(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	m := &manifest.Manifest{
		Dependencies: manifest.Dependencies{
			Requires: []*manifest.Dependency{
				{
					Name: "database",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/mysql:v1.0.0",
					},
				},
				{
					Name: "app",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/webapp:v1.0.0",
					},
					Parameters: map[string]string{
						"db-connstr": "${bundle.dependencies.database.outputs.connstr}",
					},
				},
			},
		},
	}

	graph, err := p.Porter.buildDependencyGraph(m)
	require.NoError(t, err)

	err = graph.computeExecutionOrder()
	require.NoError(t, err)

	// database must come before app
	assert.Equal(t, []string{"database", "app"}, graph.ExecutionOrder)
}

func TestComputeExecutionOrder_Diamond(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	m := &manifest.Manifest{
		Dependencies: manifest.Dependencies{
			Requires: []*manifest.Dependency{
				{
					Name: "infra",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/infra:v1.0.0",
					},
				},
				{
					Name: "database",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/mysql:v1.0.0",
					},
					Parameters: map[string]string{
						"vpc-id": "${bundle.dependencies.infra.outputs.vpc-id}",
					},
				},
				{
					Name: "cache",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/redis:v1.0.0",
					},
					Parameters: map[string]string{
						"vpc-id": "${bundle.dependencies.infra.outputs.vpc-id}",
					},
				},
				{
					Name: "app",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/webapp:v1.0.0",
					},
					Parameters: map[string]string{
						"db-connstr":    "${bundle.dependencies.database.outputs.connstr}",
						"cache-connstr": "${bundle.dependencies.cache.outputs.connstr}",
					},
				},
			},
		},
	}

	graph, err := p.Porter.buildDependencyGraph(m)
	require.NoError(t, err)

	err = graph.computeExecutionOrder()
	require.NoError(t, err)

	// Verify the execution order is valid
	assert.Len(t, graph.ExecutionOrder, 4)

	// infra must come first
	assert.Equal(t, "infra", graph.ExecutionOrder[0])

	// app must come last
	assert.Equal(t, "app", graph.ExecutionOrder[3])

	// database and cache can be in either order, but both after infra
	middleTwo := graph.ExecutionOrder[1:3]
	assert.Contains(t, middleTwo, "database")
	assert.Contains(t, middleTwo, "cache")
}

func TestComputeExecutionOrder_NoDependencies(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	m := &manifest.Manifest{
		Dependencies: manifest.Dependencies{
			Requires: []*manifest.Dependency{
				{
					Name: "database",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/mysql:v1.0.0",
					},
				},
				{
					Name: "cache",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/redis:v1.0.0",
					},
				},
			},
		},
	}

	graph, err := p.Porter.buildDependencyGraph(m)
	require.NoError(t, err)

	err = graph.computeExecutionOrder()
	require.NoError(t, err)

	// Both should be in the execution order
	assert.Len(t, graph.ExecutionOrder, 2)
	assert.Contains(t, graph.ExecutionOrder, "database")
	assert.Contains(t, graph.ExecutionOrder, "cache")
}

func TestComputeExecutionOrder_CircularDependency(t *testing.T) {
	// Manually construct a graph with a cycle since we can't create this from a manifest
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
	}

	nodeA := &DependencyNode{Name: "a", OutputsUsed: make(map[string]OutputReference)}
	nodeB := &DependencyNode{Name: "b", OutputsUsed: make(map[string]OutputReference)}
	nodeC := &DependencyNode{Name: "c", OutputsUsed: make(map[string]OutputReference)}

	// Create cycle: a -> b -> c -> a
	nodeA.Dependencies = []*DependencyNode{nodeB}
	nodeB.Dependents = []*DependencyNode{nodeA}

	nodeB.Dependencies = []*DependencyNode{nodeC}
	nodeC.Dependents = []*DependencyNode{nodeB}

	nodeC.Dependencies = []*DependencyNode{nodeA}
	nodeA.Dependents = []*DependencyNode{nodeC}

	graph.Nodes["a"] = nodeA
	graph.Nodes["b"] = nodeB
	graph.Nodes["c"] = nodeC

	err := graph.computeExecutionOrder()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestDetectCycles_SimpleCycle(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
	}

	nodeA := &DependencyNode{Name: "a", OutputsUsed: make(map[string]OutputReference)}
	nodeB := &DependencyNode{Name: "b", OutputsUsed: make(map[string]OutputReference)}

	// Create cycle: a -> b -> a
	nodeA.Dependencies = []*DependencyNode{nodeB}
	nodeB.Dependencies = []*DependencyNode{nodeA}

	graph.Nodes["a"] = nodeA
	graph.Nodes["b"] = nodeB

	err := graph.detectCycles()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestDetectCycles_SelfReference(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
	}

	nodeA := &DependencyNode{Name: "a", OutputsUsed: make(map[string]OutputReference)}
	nodeA.Dependencies = []*DependencyNode{nodeA}

	graph.Nodes["a"] = nodeA

	err := graph.detectCycles()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestDetectCycles_NoCycle(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
	}

	nodeA := &DependencyNode{Name: "a", OutputsUsed: make(map[string]OutputReference)}
	nodeB := &DependencyNode{Name: "b", OutputsUsed: make(map[string]OutputReference)}
	nodeC := &DependencyNode{Name: "c", OutputsUsed: make(map[string]OutputReference)}

	// Create chain: a -> b -> c
	nodeA.Dependencies = []*DependencyNode{nodeB}
	nodeB.Dependencies = []*DependencyNode{nodeC}

	graph.Nodes["a"] = nodeA
	graph.Nodes["b"] = nodeB
	graph.Nodes["c"] = nodeC

	err := graph.detectCycles()
	assert.NoError(t, err)
}

func TestValidateOutputReferences(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	m := &manifest.Manifest{
		Dependencies: manifest.Dependencies{
			Requires: []*manifest.Dependency{
				{
					Name: "database",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/mysql:v1.0.0",
					},
				},
				{
					Name: "app",
					Bundle: manifest.BundleCriteria{
						Reference: "getporter.org/webapp:v1.0.0",
					},
					Parameters: map[string]string{
						"db-connstr": "${bundle.dependencies.database.outputs.connstr}",
					},
				},
			},
		},
	}

	graph, err := p.Porter.buildDependencyGraph(m)
	require.NoError(t, err)

	err = graph.validateOutputReferences(m)
	assert.NoError(t, err)
}
