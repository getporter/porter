//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInspectRecursiveDependencies verifies that porter inspect --show-dependencies
// correctly displays nested (transitive) dependencies
func TestInspectRecursiveDependencies(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	// Publish mysql bundle (leaf dependency)
	publishMySQLForInspect(ctx, p)

	// Publish wordpress bundle (depends on mysql)
	publishWordpressForInspect(ctx, p)

	// Create and publish top-level bundle that depends on wordpress
	publishTopLevelBundle(ctx, p)

	// Inspect the top-level bundle with --show-dependencies
	opts := porter.ExplainOpts{}
	opts.Reference = "localhost:5000/top-level:v0.1.0"
	opts.ShowDependencies = true
	opts.MaxDependencyDepth = 10

	err := opts.Validate(nil, p.Context)
	require.NoError(t, err)

	// Capture output
	inspectOutput, err := p.GetInspectOutput(ctx, opts)
	require.NoError(t, err)

	// Verify that we have dependencies
	require.NotNil(t, inspectOutput.Dependencies)
	require.Len(t, inspectOutput.Dependencies, 1, "should have 1 direct dependency (wordpress)")

	// Verify wordpress dependency
	wordpressDep := inspectOutput.Dependencies[0]
	assert.Equal(t, "wordpress", wordpressDep.Alias)
	assert.Equal(t, "localhost:5000/wordpress:v0.1.4", wordpressDep.Reference)
	assert.Equal(t, 0, wordpressDep.Depth)

	// Verify nested mysql dependency
	require.NotNil(t, wordpressDep.Dependencies)
	require.Len(t, wordpressDep.Dependencies, 1, "wordpress should have 1 dependency (mysql)")

	mysqlDep := wordpressDep.Dependencies[0]
	assert.Equal(t, "mysql", mysqlDep.Alias)
	assert.Equal(t, "localhost:5000/mysql:v0.1.4", mysqlDep.Reference)
	assert.Equal(t, 1, mysqlDep.Depth, "mysql should be at depth 1")
}

// TestInspectDependenciesJSON verifies JSON output includes nested dependencies
func TestInspectDependenciesJSON(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	// Publish bundles
	publishMySQLForInspect(ctx, p)
	publishWordpressForInspect(ctx, p)

	// Inspect wordpress with JSON output
	opts := porter.ExplainOpts{}
	opts.Reference = "localhost:5000/wordpress:v0.1.4"
	opts.ShowDependencies = true
	opts.MaxDependencyDepth = 10
	opts.Format = "json"

	err := opts.Validate(nil, p.Context)
	require.NoError(t, err)

	output, err := p.GetInspectOutput(ctx, opts)
	require.NoError(t, err)

	// Marshal to JSON and verify structure
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	require.NoError(t, err)

	// Verify JSON contains nested dependencies
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	deps, ok := result["dependencies"].([]interface{})
	require.True(t, ok, "dependencies should be present in JSON")
	require.Len(t, deps, 1, "should have 1 direct dependency")

	// Verify the structure is correct
	dep := deps[0].(map[string]interface{})
	assert.Equal(t, "mysql", dep["alias"])
	assert.Equal(t, float64(0), dep["depth"])
}

// TestInspectMaxDepth verifies max depth limiting works
func TestInspectMaxDepth(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	// Publish bundles
	publishMySQLForInspect(ctx, p)
	publishWordpressForInspect(ctx, p)
	publishTopLevelBundle(ctx, p)

	// Inspect with max depth = 1 (should only show wordpress, not mysql)
	opts := porter.ExplainOpts{}
	opts.Reference = "localhost:5000/top-level:v0.1.0"
	opts.ShowDependencies = true
	opts.MaxDependencyDepth = 1

	err := opts.Validate(nil, p.Context)
	require.NoError(t, err)

	output, err := p.GetInspectOutput(ctx, opts)
	require.NoError(t, err)

	// Should have wordpress at depth 0
	require.Len(t, output.Dependencies, 1)
	assert.Equal(t, "wordpress", output.Dependencies[0].Alias)

	// But wordpress should have no nested dependencies (stopped at max depth)
	assert.Len(t, output.Dependencies[0].Dependencies, 0, "should not traverse beyond max depth")
}

func publishMySQLForInspect(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := os.MkdirTemp("", "porter-mysql")
	require.NoError(p.T(), err)
	defer os.RemoveAll(bunDir)

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/mysql"), bunDir)
	pwd := p.Getwd()
	p.Chdir(bunDir)
	defer p.Chdir(pwd)

	publishOpts := porter.PublishOptions{}
	publishOpts.Force = true
	err = publishOpts.Validate(p.Config)
	require.NoError(p.T(), err)

	err = p.Publish(ctx, publishOpts)
	require.NoError(p.T(), err)
}

func publishWordpressForInspect(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := os.MkdirTemp("", "porter-wordpress")
	require.NoError(p.T(), err)
	defer os.RemoveAll(bunDir)

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/wordpress"), bunDir)
	pwd := p.Getwd()
	p.Chdir(bunDir)
	defer p.Chdir(pwd)

	publishOpts := porter.PublishOptions{}
	publishOpts.Force = true
	err = publishOpts.Validate(p.Config)
	require.NoError(p.T(), err)

	err = p.Publish(ctx, publishOpts)
	require.NoError(p.T(), err)
}

func publishTopLevelBundle(ctx context.Context, p *porter.TestPorter) {
	// Create a simple bundle that depends on wordpress
	bunDir, err := os.MkdirTemp("", "porter-toplevel")
	require.NoError(p.T(), err)
	defer os.RemoveAll(bunDir)

	// Create porter.yaml for top-level bundle
	manifest := `schemaVersion: 1.0.1
name: top-level
version: 0.1.0
registry: "localhost:5000"

mixins:
  - exec

dependencies:
  requires:
    - name: wordpress
      bundle:
        reference: localhost:5000/wordpress:v0.1.4
      parameters:
        wordpress-password: test-password
        namespace: test-ns

install:
  - exec:
      command: echo
      arguments:
        - "installed top-level"

uninstall:
  - exec:
      command: echo
      arguments:
        - "uninstalled top-level"
`

	err = os.WriteFile(filepath.Join(bunDir, "porter.yaml"), []byte(manifest), 0644)
	require.NoError(p.T(), err)

	pwd := p.Getwd()
	p.Chdir(bunDir)
	defer p.Chdir(pwd)

	publishOpts := porter.PublishOptions{}
	publishOpts.Force = true
	err = publishOpts.Validate(p.Config)
	require.NoError(p.T(), err)

	err = p.Publish(ctx, publishOpts)
	require.NoError(p.T(), err)
}
