//go:build integration

package integration

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	migrationhelpers "get.porter.sh/porter/pkg/storage/migrations/testhelpers"
	testhelpers "get.porter.sh/porter/pkg/test"
	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
)

// Do a migration. This also checks for any problems with our
// connection handling which can result in panics :-)
func TestMigration(t *testing.T) {
	t.Parallel()

	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	testdata := filepath.Join(test.RepoRoot, "tests/integration/testdata/migration/")

	// Set up a PORTER_HOME with v0.38 data
	oldCfg := migrationhelpers.CreateLegacyPorterHome(t)
	defer oldCfg.Close()
	oldHome, err := oldCfg.GetHomeDir()

	// Migrate their data
	destNamespace := "migrated"
	_, output := test.RequirePorter("storage", "migrate", "--old-home", oldHome, "--old-account=src", "--namespace", destNamespace)

	// Verify that the installations were migrated to the specified namespace
	output, _ = test.RequirePorter("list", "--namespace", destNamespace, "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "installations-list-output.json"), output)

	// Verify that all the previous runs were migrated
	output, _ = test.RequirePorter("installation", "runs", "list", "--namespace", destNamespace, "hello1", "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "runs-list-hello1-output.json"), output)

	output, _ = test.RequirePorter("installation", "runs", "list", "--namespace", destNamespace, "hello-llama", "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "runs-list-hello-llama-output.json"), output)

	output, _ = test.RequirePorter("installation", "runs", "list", "--namespace", destNamespace, "creds-tutorial", "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "runs-list-creds-tutorial-output.json"), output)

	output, _ = test.RequirePorter("installation", "runs", "list", "--namespace", destNamespace, "sensitive-data", "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "runs-list-sensitive-data-output.json"), output)

	// Verify that outputs were migrated, all the installations except sensitive-data only have logs (which aren't printed by installation outputs list)
	// Show the logs from installing hello1
	output, _ = test.RequirePorter("installation", "logs", "show", "--namespace", destNamespace, "-r=01G1VJGY43HT3KZN82DS6DDPWK")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "logs-install-hello1.txt"), output)

	// Show the logs from the last run of hello-llama
	output, _ = test.RequirePorter("installation", "logs", "show", "--namespace", destNamespace, "-i=hello-llama")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "logs-hello-llama.txt"), output)

	// Show the outputs of the sensitive-data bundle
	output, _ = test.RequirePorter("installation", "outputs", "list", "--namespace", destNamespace, "sensitive-data", "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "outputs-list-sensitive-data-output.json"), output)

	// Dump out the migrated installations and make sure that all the fields are set correctly
	output, _ = test.RequirePorter("installation", "show", "--namespace", destNamespace, "hello1", "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "installation-show-hello1-output.json"), output)

	output, _ = test.RequirePorter("installation", "show", "--namespace", destNamespace, "hello-llama", "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "installation-show-hello-llama-output.json"), output)

	output, _ = test.RequirePorter("installation", "show", "--namespace", destNamespace, "creds-tutorial", "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "installation-show-creds-tutorial-output.json"), output)

	output, _ = test.RequirePorter("installation", "show", "--namespace", destNamespace, "sensitive-data", "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "installation-show-sensitive-data-output.json"), output)

	// Verify that the sensitive-data installation stored sensitive values in the secret store
	secretsDir := filepath.Join(test.PorterHomeDir, "secrets")
	secretOutput, err := ioutil.ReadFile(filepath.Join(secretsDir, "01G6K8CZ08T78WXTJYHR0NTYBS-name"))
	require.NoError(t, err, "Failed to read the secrets file for the sensitive output: name")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "secrets/01G6K8CZ08T78WXTJYHR0NTYBS-name"), string(secretOutput))
	secretParam, err := ioutil.ReadFile(filepath.Join(secretsDir, "01G6K8CZ08T78WXTJYHR0NTYBS-password"))
	require.NoError(t, err, "Failed to read the secrets file for the sensitive parameter: password")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "secrets/01G6K8CZ08T78WXTJYHR0NTYBS-password"), string(secretParam))

	// Verify that the parameter sets were migrated to the specified namespace
	output, _ = test.RequirePorter("parameters", "list", "--namespace", destNamespace, "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "parameters-list-output.json"), output)

	// Dump out the migrated parameter sets and make sure that all the fields are set correctly
	output, _ = test.RequirePorter("parameters", "show", "--namespace", destNamespace, "hello-llama", "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "parameters-show-hello-llama-output.json"), output)

	// Verify that the credential sets were migrated to the specified namespace
	output, _ = test.RequirePorter("credentials", "list", "--namespace", destNamespace, "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "credentials-list-output.json"), output)

	// Dump out the migrated credential sets and make sure that all the fields are set correctly
	output, _ = test.RequirePorter("credentials", "show", "--namespace", destNamespace, "credentials-tutorial", "--output=json")
	testhelpers.CompareGoldenFile(t, filepath.Join(testdata, "credentials-show-credentials-tutorial-output.json"), output)
}
