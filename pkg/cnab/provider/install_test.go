package cnabprovider

import (
	"testing"

	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_Install(t *testing.T) {
	t.Parallel()

	r := NewTestRuntime(t)
	r.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "bundle.json")

	args := ActionArguments{
		Action:       claim.ActionInstall,
		Installation: "mybuns",
		BundlePath:   "bundle.json",
	}
	err := r.Execute(args)
	require.NoError(t, err, "Install failed")

	c, err := r.claims.ReadLastClaim(args.Installation)
	require.NoError(t, err, "ReadLastClaim failed")

	assert.Equal(t, claim.ActionInstall, c.Action, "wrong action recorded")
	assert.Equal(t, args.Installation, c.Installation, "wrong installation name recorded")
}
