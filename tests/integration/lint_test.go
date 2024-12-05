//go:build integration

package integration

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
	"github.com/uwu-tools/magex/shx"
)

func TestLint(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// When a mixin doesn't support lint, we should not print that to the console unless we are in debug mode
	_, output, _ := test.RunPorterWith(func(cmd *shx.PreparedCommand) {
		cmd.Args("lint")
		// mybuns uses the testmixin which doesn't support lint
		cmd.In(filepath.Join(test.RepoRoot, "tests/testdata/mybuns"))
		// change verbosity to debug so that we see the error
		cmd.Env("PORTER_VERBOSITY=debug")
	})
	require.Contains(t, output, "unknown command", "an unsupported mixin command should print to the console in debug")

	_, output, _ = test.RunPorterWith(func(cmd *shx.PreparedCommand) {
		cmd.Args("lint")
		// mybuns uses the testmixin which doesn't support lint
		cmd.In(filepath.Join(test.RepoRoot, "tests/testdata/mybuns"))
		// errors are printed at the debug level to bump it up to info
		cmd.Env("PORTER_VERBOSITY=info")
	})
	require.NotContains(t, output, "unknown command", "an unsupported mixin command should not be printed to the console in info")
}

func TestLint_ApplyToParam(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	_, output, _ := test.RunPorterWith(func(cmd *shx.PreparedCommand) {
		cmd.Args("lint")
		cmd.In(filepath.Join(test.RepoRoot, "tests/integration/testdata/bundles/bundle-with-param-apply-lint-error"))
	})
	require.Contains(t, output, "error(porter-101) - Parameter does not apply to action", "parameters being used in actions to which they don't apply should be an error")
}

func TestLint_DependenciesSameName(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	_, output, _ := test.RunPorterWith(func(cmd *shx.PreparedCommand) {
		cmd.Args("lint")
		cmd.In(filepath.Join(test.RepoRoot, "tests/integration/testdata/bundles/bundle-with-samenamedeps-lint-error"))
	})
	require.Contains(t, output, "error(porter-102) - Dependency error", "multiple dependencies with the same name should be an error")
}
