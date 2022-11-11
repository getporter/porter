//go:build smoke

package smoke

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

// Test desired state workflows used by the porter operator
func TestDesiredState(t *testing.T) {
	// I am always using require, so that we stop immediately upon an error
	// A long test is hard to debug when it fails in the middle and keeps going
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	test.PrepareTestBundle()
	test.Chdir(test.TestDir)

	// Import some creds and params for mybuns
	test.RequirePorter("parameters", "apply", filepath.Join(test.RepoRoot, "tests/testdata/params/mybuns.yaml"), "--namespace=")
	test.RequirePorter("credentials", "apply", filepath.Join(test.RepoRoot, "tests/testdata/creds/mybuns.yaml"), "--namespace=")
	test.RequirePorter("credentials", "apply", filepath.Join(test.RepoRoot, "tests/testdata/creds/alt-mybuns.yaml"), "--namespace=")
	mgx.Must(shx.Copy(filepath.Join(test.RepoRoot, "tests/testdata/installations/mybuns.yaml"), "mybuns.yaml"))

	t.Run("apply installation with invalid schema", func(t *testing.T) {
		// Try to import an installation with an invalid schema
		_, _, err = test.RunPorter("installation", "apply", filepath.Join(test.RepoRoot, "tests/testdata/installations/invalid-schema.yaml"))
		require.Error(t, err, "apply should have failed because the schema of the imported document is incorrect")
		require.Contains(t, err.Error(), "invalid installation")
	})

	t.Run("apply credential set with invalid schema", func(t *testing.T) {
		// Try to import a credential set with an invalid schema
		_, _, err = test.RunPorter("credentials", "apply", filepath.Join(test.RepoRoot, "tests/testdata/creds/invalid-schema.yaml"))
		require.Error(t, err, "apply should have failed because the schema of the imported document is incorrect")
		require.Contains(t, err.Error(), "invalid credential set")
	})

	t.Run("apply parameter set with invalid schema", func(t *testing.T) {
		// Try to import a parameter set with an invalid schema
		_, _, err = test.RunPorter("parameters", "apply", filepath.Join(test.RepoRoot, "tests/testdata/params/invalid-schema.yaml"))
		require.Error(t, err, "apply should have failed because the schema of the imported document is incorrect")
		require.Contains(t, err.Error(), "invalid parameter set")
	})

	t.Run("apply new installation with uninstalled=true", func(t *testing.T) {
		// Import an installation with uninstalled=true, should do nothing
		test.EditYaml("mybuns.yaml", func(yq *yaml.Editor) error {
			return yq.SetValue("uninstalled", "true")
		})
		_, output, err := test.RunPorter("installation", "apply", "mybuns.yaml", "--namespace", "operator")
		require.NoError(t, err)
		require.Contains(t, output, "Ignoring because installation.uninstalled is true but the installation doesn't exist yet")
	})

	t.Run("apply new installation with uninstalled=false", func(t *testing.T) {
		// Now set uninstalled = false so that it's installed for the first time
		test.EditYaml("mybuns.yaml", func(yq *yaml.Editor) error {
			return yq.SetValue("uninstalled", "false")
		})
	})

	t.Run("import installation into namespace", func(t *testing.T) {
		// Import an installation, since the file is missing a namespace, it should use the --namespace flag value
		// This also tests out that --allow-docker-host-access is being defaulted properly from the Porter config file
		_, output, err := test.RunPorter("installation", "apply", "mybuns.yaml", "--namespace", "operator")
		require.NoError(t, err)
		require.Contains(t, output, "The installation is out-of-sync, running the install action")
		require.Contains(t, output, "Triggering because the installation has not completed successfully yet")
		installation := test.RequireInstallationExists("operator", "mybuns")
		require.Equal(t, "succeeded", installation.Status.ResultStatus)
	})

	t.Run("apply installation unchanged installation should not execute", func(t *testing.T) {
		// Repeat the apply command, there should be no changes detected. Using dry run because we just want to know if it _would_ be re-executed.
		_, output, err := test.RunPorter("installation", "apply", "mybuns.yaml", "--namespace", "operator", "--dry-run")
		require.NoError(t, err)
		tests.RequireOutputContains(t, output, "The installation is already up-to-date")
	})

	t.Run("apply installation with force triggers execution", func(t *testing.T) {
		// Repeat the apply command with --force, even though there are no changes, this should trigger an upgrade.
		_, output, err := test.RunPorter("installation", "apply", "mybuns.yaml", "--namespace", "operator", "--dry-run", "--force")
		require.NoError(t, err)
		tests.RequireOutputContains(t, output, "The installation is up-to-date but will be re-applied because --force was specified")
	})

	t.Run("apply installation with changed label", func(t *testing.T) {
		// Edit the installation file with a minor change that shouldn't trigger reconciliation
		test.EditYaml("mybuns.yaml", func(yq *yaml.Editor) error {
			return yq.SetValue("labels.thing", "2")
		})
		_, output, err := test.RunPorter("installation", "apply", "mybuns.yaml", "--namespace", "operator")
		require.NoError(t, err)
		tests.RequireOutputContains(t, output, "The installation is already up-to-date")
	})

	t.Run("apply installation with different parameter value", func(t *testing.T) {
		// Change a bundle parameter and trigger an upgrade
		test.EditYaml("mybuns.yaml", func(yq *yaml.Editor) error {
			return yq.SetValue("parameters.log_level", "3")
		})
		_, output, err := test.RunPorter("installation", "apply", "mybuns.yaml", "--namespace", "operator")
		require.NoError(t, err)
		tests.RequireOutputContains(t, output, "The installation is out-of-sync, running the upgrade action")

		// Validate the parameter change worked
		displayInstallation, err := test.ShowInstallation("operator", "mybuns")
		require.NoError(t, err)
		require.Equal(t, float64(3), displayInstallation.Parameters["log_level"])
	})

	t.Run("apply installation with different credential set", func(t *testing.T) {
		// Switch credentials and trigger an upgrade
		test.EditYaml("mybuns.yaml", func(yq *yaml.Editor) error {
			return yq.SetValue("credentialSets[0]", "alt-mybuns")
		})
		_, output, err := test.RunPorter("installation", "apply", "mybuns.yaml", "--namespace", "operator")
		require.NoError(t, err)
		tests.RequireOutputContains(t, output, "The installation is out-of-sync, running the upgrade action")
	})

	t.Run("apply installation with uninstalled=true", func(t *testing.T) {
		// Uninstall by setting uninstalled: true
		test.EditYaml("mybuns.yaml", func(yq *yaml.Editor) error {
			return yq.SetValue("uninstalled", "true")
		})
		_, output, err := test.RunPorter("installation", "apply", "mybuns.yaml", "--namespace", "operator")
		require.NoError(t, err)
		tests.RequireOutputContains(t, output, "The installation is out-of-sync, running the uninstall action")
	})
}
