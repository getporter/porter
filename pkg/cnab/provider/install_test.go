package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_Install(t *testing.T) {
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Teardown()

	r.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "bundle.json")

	args := ActionArguments{
		Namespace:    "dev",
		Action:       cnab.ActionInstall,
		Installation: "mybuns",
		BundlePath:   "bundle.json",
	}
	err := r.Execute(args)
	require.NoError(t, err, "Install failed")

	c, err := r.claims.GetLastRun(args.Namespace, args.Installation)
	require.NoError(t, err, "GetLastRun failed")

	assert.Equal(t, cnab.ActionInstall, c.Action, "wrong action recorded")
	assert.Equal(t, args.Installation, c.Installation, "wrong installation name recorded")
}
