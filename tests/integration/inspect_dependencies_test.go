//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/yaml"
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

func publishWordpressWithMissingDependency(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := os.MkdirTemp("", "porter-wordpress-fail")
	require.NoError(p.T(), err)
	defer os.RemoveAll(bunDir)

	// Create porter.yaml with non-existent mysql dependency
	manifest := `schemaVersion: 1.0.1
name: wordpress-fail-test
version: 0.1.0
registry: "localhost:5000"

mixins:
  - exec

dependencies:
  requires:
    - name: mysql
      bundle:
        reference: localhost:5000/mysql-nonexistent:v99.99.99

install:
  - exec:
      command: echo
      arguments:
        - "installed wordpress"

uninstall:
  - exec:
      command: echo
      arguments:
        - "uninstalled wordpress"
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

// TestInspectDiamondDependencyIsResolvedOnce verifies that a dependency
// shared by two different parents (a diamond: diamond-top requires both
// diamond-app-a and diamond-app-b, and both of those require the same
// mysql bundle) is resolved consistently under each branch -- same
// reference, correct depth, no resolution errors -- confirming transitive
// resolution reaches the shared dependency correctly from both paths.
func TestInspectDiamondDependencyIsResolvedOnce(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	publishMySQLForInspect(ctx, p)
	publishDiamondAppForInspect(ctx, p, "diamond-app-a")
	publishDiamondAppForInspect(ctx, p, "diamond-app-b")
	publishDiamondTopForInspect(ctx, p)

	opts := porter.ExplainOpts{}
	opts.Reference = "localhost:5000/diamond-top:v0.1.0"
	opts.ShowDependencies = true
	opts.MaxDependencyDepth = 10

	err := opts.Validate(nil, p.Context)
	require.NoError(t, err)

	output, err := p.GetInspectOutput(ctx, opts)
	require.NoError(t, err)

	require.Len(t, output.Dependencies, 2, "should have 2 direct dependencies (diamond-app-a, diamond-app-b)")

	deps := map[string]porter.InspectableDependency{}
	for _, dep := range output.Dependencies {
		deps[dep.Alias] = dep
	}
	appA, ok := deps["diamond-app-a"]
	require.True(t, ok, "expected a diamond-app-a dependency")
	appB, ok := deps["diamond-app-b"]
	require.True(t, ok, "expected a diamond-app-b dependency")

	assert.Equal(t, 0, appA.Depth)
	assert.Equal(t, 0, appB.Depth)

	require.Len(t, appA.Dependencies, 1, "diamond-app-a should have 1 dependency (mysql)")
	require.Len(t, appB.Dependencies, 1, "diamond-app-b should have 1 dependency (mysql)")

	mysqlUnderA := appA.Dependencies[0]
	mysqlUnderB := appB.Dependencies[0]

	assert.Equal(t, "mysql", mysqlUnderA.Alias)
	assert.Equal(t, "mysql", mysqlUnderB.Alias)
	assert.Equal(t, 1, mysqlUnderA.Depth, "mysql should be at depth 1 under diamond-app-a")
	assert.Equal(t, 1, mysqlUnderB.Depth, "mysql should be at depth 1 under diamond-app-b")
	assert.False(t, mysqlUnderA.ResolutionFailed)
	assert.False(t, mysqlUnderB.ResolutionFailed)

	require.Equal(t, "localhost:5000/mysql:v0.1.4", mysqlUnderA.Reference)
	assert.Equal(t, mysqlUnderA.Reference, mysqlUnderB.Reference,
		"a dependency shared by two parents must resolve to the same reference under both branches")
}

// publishDiamondAppForInspect publishes the diamond-app fixture bundle under
// the given name (diamond-app-a or diamond-app-b), so the same template can
// back both branches of the diamond.
func publishDiamondAppForInspect(ctx context.Context, p *porter.TestPorter, name string) {
	bunDir, err := os.MkdirTemp("", "porter-"+name)
	require.NoError(p.T(), err)
	defer os.RemoveAll(bunDir)

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/diamond-app"), bunDir)

	manifestPath := filepath.Join(bunDir, "porter.yaml")
	e := yaml.NewEditor(p.FileSystem)
	require.NoError(p.T(), e.ReadFile(manifestPath))
	require.NoError(p.T(), e.SetValue("name", name))
	require.NoError(p.T(), e.WriteFile(manifestPath))

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

func publishDiamondTopForInspect(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := os.MkdirTemp("", "porter-diamond-top")
	require.NoError(p.T(), err)
	defer os.RemoveAll(bunDir)

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/diamond-top"), bunDir)
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

// TestInspectDependenciesWithFailures verifies that failed dependencies are marked with warning emoji
func TestInspectDependenciesWithFailures(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	// Publish wordpress bundle with non-existent mysql dependency
	publishWordpressWithMissingDependency(ctx, p)

	// Inspect wordpress with --show-dependencies (mysql dep should fail to resolve)
	opts := porter.ExplainOpts{}
	opts.Reference = "localhost:5000/wordpress-fail-test:v0.1.0"
	opts.ShowDependencies = true
	opts.MaxDependencyDepth = 10

	err := opts.Validate(nil, p.Context)
	require.NoError(t, err)

	// Capture output
	output, err := p.GetInspectOutput(ctx, opts)
	require.NoError(t, err)

	// Verify that we have dependencies
	require.NotNil(t, output.Dependencies)
	require.Len(t, output.Dependencies, 1, "should have 1 direct dependency (mysql)")

	// Verify mysql dependency failed
	mysqlDep := output.Dependencies[0]
	assert.Equal(t, "mysql", mysqlDep.Alias)
	assert.Equal(t, "localhost:5000/mysql-nonexistent:v99.99.99", mysqlDep.Reference)
	assert.True(t, mysqlDep.ResolutionFailed, "mysql dependency should have failed to resolve")
	assert.NotEmpty(t, mysqlDep.ResolutionError, "should have error message")
	assert.Len(t, mysqlDep.Dependencies, 0, "failed dependency should have no nested dependencies")
}

// TestInspectWiringEdgeIsResolved verifies that porter inspect --show-dependencies
// detects a DependenciesV2 wiring reference: wiring-top requires both "infra"
// and "app", and app's "connstr" parameter is wired from infra's "ip"
// output. This checks that the wiring edge is correctly resolved from the
// published bundle.json (not just from an in-memory manifest), proving
// output-wiring is part of what "resolved correctly" means for a
// transitively-published dependency graph.
func TestInspectWiringEdgeIsResolved(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.Config.SetExperimentalFlags(experimental.FlagDependenciesV2)

	publishWiringInfraForInspect(ctx, p)
	publishWiringAppForInspect(ctx, p)
	publishWiringTopForInspect(ctx, p)

	opts := porter.ExplainOpts{}
	opts.Reference = "localhost:5000/wiring-top:v0.1.0"
	opts.ShowDependencies = true
	opts.MaxDependencyDepth = 10

	err := opts.Validate(nil, p.Context)
	require.NoError(t, err)

	output, err := p.GetInspectOutput(ctx, opts)
	require.NoError(t, err)

	require.Len(t, output.Dependencies, 2, "should have 2 direct dependencies (infra, app)")

	deps := map[string]porter.InspectableDependency{}
	for _, dep := range output.Dependencies {
		deps[dep.Alias] = dep
	}
	infra, ok := deps["infra"]
	require.True(t, ok, "expected an infra dependency")
	app, ok := deps["app"]
	require.True(t, ok, "expected an app dependency")

	assert.Equal(t, 0, infra.Depth)
	assert.Equal(t, 0, app.Depth)
	assert.Empty(t, infra.WiringEdges, "infra doesn't wire from anything")

	require.Len(t, app.WiringEdges, 1, "app's connstr parameter should be wired from infra's output")
	wiring := app.WiringEdges[0]
	assert.Equal(t, "parameters", wiring.Field)
	assert.Equal(t, "connstr", wiring.FieldName)
	assert.Equal(t, "infra", wiring.SourceDependencyAlias)
	assert.Equal(t, "ip", wiring.SourceOutput)
}

func publishWiringInfraForInspect(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := os.MkdirTemp("", "porter-wiring-infra")
	require.NoError(p.T(), err)
	defer os.RemoveAll(bunDir)

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/wiring-infra"), bunDir)
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

func publishWiringAppForInspect(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := os.MkdirTemp("", "porter-wiring-app")
	require.NoError(p.T(), err)
	defer os.RemoveAll(bunDir)

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/wiring-app"), bunDir)
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

func publishWiringTopForInspect(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := os.MkdirTemp("", "porter-wiring-top")
	require.NoError(p.T(), err)
	defer os.RemoveAll(bunDir)

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/wiring-top"), bunDir)
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
