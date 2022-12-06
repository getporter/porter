//go:build integration

package integration

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
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
