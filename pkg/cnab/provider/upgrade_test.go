package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_Upgrade(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		defer r.Teardown()

		r.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "bundle.json")

		i := r.TestClaims.CreateInstallation(claims.NewInstallation("", "mybuns"))
		r.TestClaims.CreateRun(i.NewRun(cnab.ActionInstall))

		args := ActionArguments{
			Action:       cnab.ActionUpgrade,
			Installation: "mybuns",
			BundlePath:   "bundle.json",
		}
		err := r.Execute(args)
		require.NoError(t, err, "Upgrade failed")

		c, err := r.claims.GetLastRun(args.Namespace, args.Installation)
		require.NoError(t, err, "GetLastRun failed")

		assert.Equal(t, cnab.ActionUpgrade, c.Action, "wrong action recorded")
		assert.Equal(t, args.Installation, c.Installation, "wrong installation name recorded")
	})

	t.Run("requires existing claim", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		defer r.Teardown()

		r.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "bundle.json")

		args := ActionArguments{
			Action:       cnab.ActionUpgrade,
			Installation: "mybuns",
			BundlePath:   "bundle.json",
		}
		err := r.Execute(args)
		require.ErrorIs(t, err, storage.ErrNotFound{})
	})
}
