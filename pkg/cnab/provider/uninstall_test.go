package cnabprovider

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"

	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_Uninstall(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r := NewTestRuntime(t)
		r.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "bundle.json")

		existingClaim, err := claim.New("mybuns", claim.ActionInstall, bundle.Bundle{}, nil)
		require.NoError(t, err, "New claim failed")
		err = r.claims.SaveClaim(existingClaim)
		require.NoError(t, err, "SaveClaim failed")

		args := ActionArguments{
			Action:       claim.ActionUninstall,
			Installation: "mybuns",
			BundlePath:   "bundle.json",
		}
		err = r.Execute(args)
		require.NoError(t, err, "Uninstall failed")

		c, err := r.claims.ReadLastClaim(args.Installation)
		require.NoError(t, err, "ReadLastClaim failed")

		assert.Equal(t, claim.ActionUninstall, c.Action, "wrong action recorded")
		assert.Equal(t, args.Installation, c.Installation, "wrong installation name recorded")
	})

	t.Run("requires existing claim", func(t *testing.T) {
		r := NewTestRuntime(t)
		r.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "bundle.json")

		args := ActionArguments{
			Action:       claim.ActionUninstall,
			Installation: "mybuns",
			BundlePath:   "bundle.json",
		}
		err := r.Execute(args)
		require.Error(t, err, "Uninstall should have failed")
		assert.Contains(t, err.Error(), "Installation does not exist")
	})
}
