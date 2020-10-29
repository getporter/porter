package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

func TestNewDisplayInstallation(t *testing.T) {
	bun := bundle.Bundle{Name: "wordpress"}

	t.Run("install exists", func(t *testing.T) {
		cp := claims.NewTestClaimProvider(t)
		install := cp.CreateClaim("wordpress", claim.ActionInstall, bun, nil)
		cp.CreateResult(install, claim.StatusSucceeded)

		upgrade := cp.CreateClaim("wordpress", claim.ActionUpgrade, bun, nil)
		cp.CreateResult(upgrade, claim.StatusRunning)

		i, err := cp.ReadInstallation("wordpress")
		require.NoError(t, err, "ReadInstallation failed")

		di, err := NewDisplayInstallation(i)
		require.NoError(t, err, "NewDisplayInstallation failed")

		require.Equal(t, di.Name, i.Name, "invalid installation name")
		require.True(t, install.Created.Equal(di.Created), "invalid created time")
		require.True(t, upgrade.Created.Equal(di.Modified), "invalid modified time")
		require.Equal(t, claim.ActionUpgrade, di.Action, "invalid last action")
		require.Equal(t, claim.StatusRunning, di.Status, "invalid last status")
	})

	t.Run("install does not exist", func(t *testing.T) {
		cp := claims.NewTestClaimProvider(t)
		dryRun := cp.CreateClaim("wordpress", "dry-run", bun, nil)
		cp.CreateResult(dryRun, claim.StatusSucceeded)

		i, err := cp.ReadInstallation("wordpress")
		require.NoError(t, err, "ReadInstallation failed")

		di, err := NewDisplayInstallation(i)
		require.NoError(t, err, "NewDisplayInstallation failed")

		require.Equal(t, di.Name, i.Name, "invalid installation name")
		require.True(t, dryRun.Created.Equal(di.Created), "invalid created time")
		require.True(t, dryRun.Created.Equal(di.Modified), "invalid modified time")
		require.Equal(t, "dry-run", di.Action, "invalid last action")
		require.Equal(t, claim.StatusSucceeded, di.Status, "invalid last status")
	})
}
