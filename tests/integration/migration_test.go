//go:build integration

package integration

import (
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

	// Set up a PORTER_HOME with v0.38 data
	oldCfg := migrationhelpers.CreateLegacyPorterHome(t)
	defer oldCfg.Close()
	oldHome, err := oldCfg.GetHomeDir()

	// Migrate their data
	destNamespace := "migrated"
	_, output := test.RequirePorter("storage", "migrate", "--old-home", oldHome, "--old-account=src", "--namespace", destNamespace)

	// Verify that the installations were migrated to the specified namespace
	output, _ = test.RequirePorter("list", "--namespace", destNamespace)
	testhelpers.CompareGoldenFile(t, "testdata/migration/installations-list-output.txt", output)

	// Verify that all the previous runs were migrated
	output, _ = test.RequirePorter("installation", "runs", "list", "--namespace", destNamespace, "hello1")
	testhelpers.CompareGoldenFile(t, "testdata/migration/runs-list-hello1-output.txt", output)

	output, _ = test.RequirePorter("installation", "runs", "list", "--namespace", destNamespace, "hello-llama")
	testhelpers.CompareGoldenFile(t, "testdata/migration/runs-list-hello-llama-output.txt", output)

	output, _ = test.RequirePorter("installation", "runs", "list", "--namespace", destNamespace, "creds-tutorial")
	testhelpers.CompareGoldenFile(t, "testdata/migration/runs-list-creds-tutorial-output.txt", output)

	// Dump out the migrated installations and make sure that all the fields are set correctly
	output, _ = test.RequirePorter("installation", "show", "--namespace", destNamespace, "hello1", "--output=json")
	testhelpers.CompareGoldenFile(t, "testdata/migration/installation-show-hello1-output.txt", output)

	output, _ = test.RequirePorter("installation", "show", "--namespace", destNamespace, "hello-llama", "--output=json")
	testhelpers.CompareGoldenFile(t, "testdata/migration/installation-show-hello-llama-output.txt", output)

	output, _ = test.RequirePorter("installation", "show", "--namespace", destNamespace, "creds-tutorial", "--output=json")
	testhelpers.CompareGoldenFile(t, "testdata/migration/installation-show-creds-tutorial-output.txt", output)

	// Verify that the parameter sets were migrated to the specified namespace
	output, _ = test.RequirePorter("parameters", "list", "--namespace", destNamespace)
	testhelpers.CompareGoldenFile(t, "testdata/migration/parameters-list-output.txt", output)

	// Dump out the migrated parameter sets and make sure that all the fields are set correctly
	output, _ = test.RequirePorter("parameters", "show", "--namespace", destNamespace, "hello-llama", "--output=json")
	testhelpers.CompareGoldenFile(t, "testdata/migration/parameters-show-hello-llama-output.txt", output)

	// Verify that the credential sets were migrated to the specified namespace
	output, _ = test.RequirePorter("credentials", "list", "--namespace", destNamespace)
	testhelpers.CompareGoldenFile(t, "testdata/migration/credentials-list-output.txt", output)

	// Dump out the migrated credential sets and make sure that all the fields are set correctly
	output, _ = test.RequirePorter("credentials", "show", "--namespace", destNamespace, "credentials-tutorial", "--output=json")
	testhelpers.CompareGoldenFile(t, "testdata/migration/credentials-show-credentials-tutorial-output.txt", output)
}
