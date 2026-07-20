//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/require"
)

// publishAndInstallReconcileTestBundle builds, publishes to the local test
// registry, and installs a unique copy of the bundle-with-credential-set-upgrade
// fixture using the specified credential set. Reconciliation (porter
// installations apply) requires the installation to declare a resolvable
// bundle reference, which only happens when installing by reference rather
// than from a local directory, so this helper always installs by reference.
// Returns the installation name, which is also the unique bundle name.
func publishAndInstallReconcileTestBundle(ctx context.Context, p *porter.TestPorter, credSet string) string {
	t := p.TestConfig.TestContext.T

	bundleName := p.AddTestBundleDir("testdata/bundles/bundle-with-credential-set-upgrade", true)
	ref := fmt.Sprintf("localhost:5000/%s:v0.1.0", bundleName)

	publishOpts := porter.PublishOptions{}
	publishOpts.Reference = ref
	publishOpts.InsecureRegistry = true
	err := publishOpts.Validate(p.Config)
	require.NoError(t, err)
	err = p.Publish(ctx, publishOpts)
	require.NoError(t, err)

	installOpts := porter.NewInstallOptions()
	installOpts.Reference = ref
	installOpts.Name = bundleName
	installOpts.InsecureRegistry = true
	installOpts.CredentialIdentifiers = []string{credSet}
	err = installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)
	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err)
	p.TestConfig.TestContext.ClearOutputs()

	return bundleName
}

// TestReconcile_NoChanges_StaysUpToDate verifies that reconciling an
// installation with no changes to its bundle, parameters or credentials
// does not trigger another bundle run.
func TestReconcile_NoChanges_StaysUpToDate(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	err := p.Credentials.InsertCredentialSet(ctx, storage.NewCredentialSet("", "reconcile-creds",
		storage.ValueStrategy("token", "same-secret")))
	require.NoError(t, err)

	installationName := publishAndInstallReconcileTestBundle(ctx, p, "reconcile-creds")

	tokenUsed, err := p.ReadBundleOutput(ctx, "token-used", installationName, "")
	require.NoError(t, err, "could not read the token-used output")
	require.Equal(t, "same-secret", tokenUsed, "sanity check: install used the seeded credential value")

	installation, err := p.Installations.GetInstallation(ctx, "", installationName)
	require.NoError(t, err)
	runIDBefore := installation.Status.RunID

	reconcileOpts := porter.ReconcileOptions{
		Namespace:    "",
		Name:         installationName,
		Installation: installation,
	}
	err = p.ReconcileInstallation(ctx, reconcileOpts)
	require.NoError(t, err)

	installation, err = p.Installations.GetInstallation(ctx, "", installationName)
	require.NoError(t, err)
	require.Equal(t, runIDBefore, installation.Status.RunID,
		"no changes were made, reconciling should not have triggered another run")
}

// TestReconcile_CredentialValueChanged_TriggersUpgrade verifies the fix for
// #1781: rotating a value inside a credential set - without changing which
// named sets are attached to the installation - must trigger a new bundle
// run, not just a change to the attached credential set names. It also
// confirms the upgrade actually runs with the rotated value, not a stale one.
func TestReconcile_CredentialValueChanged_TriggersUpgrade(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	err := p.Credentials.InsertCredentialSet(ctx, storage.NewCredentialSet("", "reconcile-creds",
		storage.ValueStrategy("token", "old-secret")))
	require.NoError(t, err)

	installationName := publishAndInstallReconcileTestBundle(ctx, p, "reconcile-creds")

	tokenUsed, err := p.ReadBundleOutput(ctx, "token-used", installationName, "")
	require.NoError(t, err, "could not read the token-used output")
	require.Equal(t, "old-secret", tokenUsed, "sanity check: install used the seeded credential value")

	// Rotate the value of the credential inside "reconcile-creds". The
	// installation's attached credential set names never change.
	creds, err := p.Credentials.GetCredentialSet(ctx, "", "reconcile-creds")
	require.NoError(t, err)
	creds.Credentials = secrets.StrategyList{storage.ValueStrategy("token", "new-secret")}
	err = p.Credentials.UpdateCredentialSet(ctx, creds)
	require.NoError(t, err)

	installation, err := p.Installations.GetInstallation(ctx, "", installationName)
	require.NoError(t, err)
	require.Equal(t, []string{"reconcile-creds"}, installation.CredentialSets,
		"sanity check: the attached credential set name did not change")
	runIDBefore := installation.Status.RunID

	reconcileOpts := porter.ReconcileOptions{
		Namespace:    "",
		Name:         installationName,
		Installation: installation,
	}
	err = p.ReconcileInstallation(ctx, reconcileOpts)
	require.NoError(t, err)

	installation, err = p.Installations.GetInstallation(ctx, "", installationName)
	require.NoError(t, err)
	require.NotEqual(t, runIDBefore, installation.Status.RunID,
		"rotating the credential value should have triggered a new run even though the credential set name did not change")
	require.Equal(t, "upgrade", installation.Status.Action)

	tokenUsed, err = p.ReadBundleOutput(ctx, "token-used", installationName, "")
	require.NoError(t, err, "could not read the token-used output after reconciling")
	require.Equal(t, "new-secret", tokenUsed,
		"the upgrade triggered by reconciliation should have used the rotated credential value")
}
