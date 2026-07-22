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

// TestInspectResolvesExistingInstallation verifies that porter inspect
// --show-dependencies (#2627) resolves a v2 dependency to an
// already-installed, compatible installation instead of pulling a new
// instance of its bundle, when the dependency declares a matching sharing
// group, no bundle interface, and the candidate installation is in the same
// namespace.
func TestInspectResolvesExistingInstallation(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.Config.SetExperimentalFlags(experimental.FlagDependenciesV2)

	namespace := p.RandomString(10)

	// Publish mysql, then install it by OCI reference (not from a local
	// build context) into namespace and label it with the "myapp" sharing
	// group, matching what wordpressv2's mysql dependency declares -- so
	// it's already-installed and reusable by the time wordpress's
	// dependency graph is resolved below.
	publishMySQLV2(ctx, p)
	installMySQLByReference(ctx, p, namespace)

	publishWordpressV2(ctx, p)

	opts := porter.ExplainOpts{}
	opts.Reference = "localhost:5000/wordpress:v0.1.4"
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
	assert.Empty(t, mysqlDep.Dependencies, "a dependency resolved to an existing installation is a leaf: its own dependencies aren't re-resolved")
}

// TestInspectDoesNotResolveExistingInstallationAcrossNamespaces verifies the
// negative case: an installed, correctly labeled mysql installation in a
// different namespace than the one being inspected is not reused, so the
// dependency is pulled as a new bundle instead (#2627's namespace scoping).
func TestInspectDoesNotResolveExistingInstallationAcrossNamespaces(t *testing.T) {
	t.Parallel()
	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	p.Config.SetExperimentalFlags(experimental.FlagDependenciesV2)

	installedNamespace := p.RandomString(10)
	inspectedNamespace := p.RandomString(10)

	publishMySQLV2(ctx, p)
	installMySQLByReference(ctx, p, installedNamespace)

	publishWordpressV2(ctx, p)

	opts := porter.ExplainOpts{}
	opts.Reference = "localhost:5000/wordpress:v0.1.4"
	opts.Namespace = inspectedNamespace
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

func publishMySQLV2(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := os.MkdirTemp("", "porter-mysqlv2-")
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

// installMySQLByReference installs the published mysql bundle into namespace
// by OCI reference (localhost:5000/mysql:v0.1.4), so the resulting
// installation's tracked Bundle reference is populated (TrackBundle) --
// unlike installing from a local build context, which leaves it empty --
// making it a realistic candidate for #2627's reference-based matching. It's
// then labeled with the "myapp" sharing group, matching wordpressv2's mysql
// dependency declaration.
func installMySQLByReference(ctx context.Context, p *porter.TestPorter, namespace string) {
	installOpts := porter.NewInstallOptions()
	installOpts.Reference = "localhost:5000/mysql:v0.1.4"
	installOpts.Name = "mysql"
	installOpts.Namespace = namespace
	installOpts.CredentialIdentifiers = []string{"ci"}

	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(p.T(), err, "validation of install opts for mysql failed")

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(p.T(), err, "install of mysql failed in namespace %s", namespace)

	mysqlInst, err := p.Installations.GetInstallation(ctx, namespace, "mysql")
	require.NoError(p.T(), err, "could not fetch the mysql installation")

	mysqlInst.SetLabel("sh.porter.SharingGroup", "myapp")
	err = p.Installations.UpdateInstallation(ctx, mysqlInst)
	require.NoError(p.T(), err, "could not label the mysql installation")
}

func publishWordpressV2(ctx context.Context, p *porter.TestPorter) {
	bunDir, err := os.MkdirTemp("", "porter-wordpressv2-")
	require.NoError(p.T(), err)
	defer os.RemoveAll(bunDir)

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.RepoRoot, "build/testdata/bundles/wordpressv2"), bunDir)

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
