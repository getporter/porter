//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/testdata"
	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

// Validate that we can use PORTER_RUNTIME_DRIVER with
// porter commands and have that set the --driver flag.
func TestBindRuntimeDriverConfiguration(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	require.NoError(t, shx.Copy(filepath.Join(test.RepoRoot, "tests/testdata/installations/mybuns.yaml"), test.TestDir))
	test.Chdir(test.TestDir)

	// Set the driver to something that will fail validation so we know it was picked up
	os.Setenv("PORTER_RUNTIME_DRIVER", "fake")
	defer os.Unsetenv("PORTER_RUNTIME_DRIVER")

	// Check that the imperative commands are using this environment variable
	_, _, err = test.RunPorter("install", testdata.MyBuns)
	tests.RequireErrorContains(t, err, "unsupported driver", "install does not have --driver wired properly")

	_, _, err = test.RunPorter("upgrade", testdata.MyBuns)
	tests.RequireErrorContains(t, err, "unsupported driver", "upgrade does not have --driver wired properly")

	_, _, err = test.RunPorter("invoke", testdata.MyBuns, "--action=ping")
	tests.RequireErrorContains(t, err, "unsupported driver", "invoke does not have --driver wired properly")

	_, _, err = test.RunPorter("uninstall", testdata.MyBuns)
	tests.RequireErrorContains(t, err, "unsupported driver", "uninstall does not have --driver wired properly")

	test.PrepareTestBundle() // apply tries to pull the bundle before the driver flag is validated
	_, output, _ := test.RunPorter("installation", "apply", "mybuns.yaml")
	tests.RequireOutputContains(t, output, "unsupported driver", "apply does not have --driver wired properly")
}

// Validate that we can use PORTER_BUILD_DRIVER with
// porter build and have that set the --driver flag.
func TestBindBuildDriverConfiguration(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Set the driver to something that will fail validation so we know it was picked up
	os.Setenv("PORTER_BUILD_DRIVER", "fake")
	defer os.Unsetenv("PORTER_BUILD_DRIVER")

	t.Run("build", func(t *testing.T) {
		_, _, err = test.RunPorter("build")
		tests.RequireErrorContains(t, err, "invalid --driver")
	})
}
