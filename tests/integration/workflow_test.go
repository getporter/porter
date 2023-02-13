//go:build integration

package integration

import (
	"fmt"
	"testing"

	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
)

func TestWorkflow(t *testing.T) {
	// Since we are working with depsv2 enabled, don't reuse a bundle that was already built for depsv1 in other tests
	test, err := tester.NewTestWithConfig(t, "tests/testdata/config/config-with-depsv2.yaml")
	defer test.Close()
	require.NoError(t, err, "test setup failed")
	test.PrepareTestBundle()

	test.TestContext.AddTestFileFromRoot("tests/testdata/installations/mybuns.yaml", "mybuns.yaml")

	// First validate the plan for the workflow
	// TODO(PEP003): Do we want to use different terms/commands for generating a workflow? This pretty much associates --dry-run with "print out your workflow"
	workflowContents, output := test.RequirePorter("installation", "apply", "mybuns.yaml", "--output=yaml", "--dry-run")
	fmt.Println(output)
	// TODO(PEP003): Until we have a display workflow, this comparison doesn't work because of extra status printed out
	_ = workflowContents
	//testhelpers.CompareGoldenFile(t, "testdata/workflow/mybuns.yaml", workflowContents)

	// Run the workflow
	_, output = test.RequirePorter("installation", "apply", "mybuns.yaml")
	fmt.Println(output)

	// TODO A workflow should be persisted, and it should match the execution plan generated first with --dry-run
	// We don't expose workflow commands yet though so the only way to test this is call the db directly

	// We should have 2 installations, mybuns and mydb
	test.RequireInstallationExists(test.CurrentNamespace(), "mybuns")
	mydb := test.RequireInstallationExists(test.CurrentNamespace(), "mybuns/db")
	require.Contains(t, mydb.Parameters, "database", "mybuns should have explicitly set the database parameter on its db dependency")
	require.Equal(t, "bigdb", mydb.Parameters["database"], "incorrect value used for the database parameter on the db dependency, expected the hard coded value specified by the root bundle")

	// TODO mydb should have a parameter that was set by the workflow, e.g. the db name
	// TODO mybuns should have used an output from mydb that we saved as a root bundle output so that we can validate that it was used properly
}
