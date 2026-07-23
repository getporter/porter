//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInspectResolvesExistingInstallationWithBundleInterface verifies #2626
// end-to-end: a v2 dependency declaring a bundle interface (interface.document
// requiring the "mysql-password" output) resolves to an already-installed,
// compatible installation -- instead of always pulling, which was the only
// behavior before #2626 -- when that installation actually has the required
// output recorded.
func TestInspectResolvesExistingInstallationWithBundleInterface(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.Config.SetExperimentalFlags(experimental.FlagDependenciesV2)

	namespace := p.RandomString(10)

	publishMySQLV2(ctx, p)
	// mysql's real install writes /cnab/app/outputs/mysql-password, so the
	// resulting installation genuinely satisfies the interface below --
	// no manual output seeding needed.
	installMySQLByReference(ctx, p, namespace)

	publishWordpressV2Interface(ctx, p)

	opts := porter.ExplainOpts{}
	opts.Reference = "localhost:5000/wordpress-interface:v0.1.4"
	opts.Namespace = namespace
	opts.ShowDependencies = true
	opts.MaxDependencyDepth = 10

	err := opts.Validate(nil, p.Context)
	require.NoError(t, err)

	inspectOutput, err := p.GetInspectOutput(ctx, opts)
	require.NoError(t, err)

	require.Len(t, inspectOutput.Dependencies, 1)
	mysqlDep := inspectOutput.Dependencies[0]
	assert.Equal(t, "mysql", mysqlDep.Alias)
	assert.False(t, mysqlDep.ResolutionFailed)
	assert.Equal(t, namespace+"/mysql", mysqlDep.ResolvedInstallation)
}

// TestInspectFallsBackToPullWhenBundleInterfaceOutputMissing verifies the
// negative case for #2626: a bundle interface requiring an output mysql
// never produces (interface.document requires "mysql-connstr", which isn't
// among mysql's declared outputs) is never satisfied by any mysql
// installation, so the dependency is pulled as a new bundle instead of
// being reused.
func TestInspectFallsBackToPullWhenBundleInterfaceOutputMissing(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.Config.SetExperimentalFlags(experimental.FlagDependenciesV2)

	namespace := p.RandomString(10)

	publishMySQLV2(ctx, p)
	installMySQLByReference(ctx, p, namespace)

	publishWordpressV2InterfaceMissing(ctx, p)

	opts := porter.ExplainOpts{}
	opts.Reference = "localhost:5000/wordpress-interface-missing:v0.1.4"
	opts.Namespace = namespace
	opts.ShowDependencies = true
	opts.MaxDependencyDepth = 10

	err := opts.Validate(nil, p.Context)
	require.NoError(t, err)

	inspectOutput, err := p.GetInspectOutput(ctx, opts)
	require.NoError(t, err)

	require.Len(t, inspectOutput.Dependencies, 1)
	mysqlDep := inspectOutput.Dependencies[0]
	assert.Equal(t, "mysql", mysqlDep.Alias)
	assert.False(t, mysqlDep.ResolutionFailed)
	assert.Empty(t, mysqlDep.ResolvedInstallation)
}

// TestInspectResolvesExistingInstallationWithBundleInterfaceReference
// verifies #2626's interface.Reference path end-to-end: a v2 dependency
// whose interface is declared via a separate, published "interface bundle"
// (mysql-interface, which is never installed -- it only exists to declare
// the outputs a bundle must provide) resolves to an already-installed mysql
// installation, the same as interface.document does, but by pulling the
// referenced bundle to determine the required outputs instead of reading
// them inline.
func TestInspectResolvesExistingInstallationWithBundleInterfaceReference(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.Config.SetExperimentalFlags(experimental.FlagDependenciesV2)

	namespace := p.RandomString(10)

	publishMySQLV2(ctx, p)
	// mysql's real install writes /cnab/app/outputs/mysql-password, so the
	// resulting installation genuinely satisfies mysql-interface's declared
	// "mysql-password" output below.
	installMySQLByReference(ctx, p, namespace)

	publishBundleDir(ctx, p, "mysql-interface")
	publishBundleDir(ctx, p, "wordpressv2-interface-reference")

	opts := porter.ExplainOpts{}
	opts.Reference = "localhost:5000/wordpress-interface-reference:v0.1.4"
	opts.Namespace = namespace
	opts.ShowDependencies = true
	opts.MaxDependencyDepth = 10

	err := opts.Validate(nil, p.Context)
	require.NoError(t, err)

	inspectOutput, err := p.GetInspectOutput(ctx, opts)
	require.NoError(t, err)

	require.Len(t, inspectOutput.Dependencies, 1)
	mysqlDep := inspectOutput.Dependencies[0]
	assert.Equal(t, "mysql", mysqlDep.Alias)
	assert.False(t, mysqlDep.ResolutionFailed)
	assert.Equal(t, namespace+"/mysql", mysqlDep.ResolvedInstallation)
}

func publishWordpressV2Interface(ctx context.Context, p *porter.TestPorter) {
	publishBundleDir(ctx, p, "wordpressv2-interface")
}

func publishWordpressV2InterfaceMissing(ctx context.Context, p *porter.TestPorter) {
	publishBundleDir(ctx, p, "wordpressv2-interface-missing")
}

func publishBundleDir(ctx context.Context, p *porter.TestPorter, bundleDirName string) {
	bunDir, err := os.MkdirTemp("", "porter-"+bundleDirName+"-")
	require.NoError(p.T(), err)
	defer os.RemoveAll(bunDir)

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles", bundleDirName), bunDir)

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
