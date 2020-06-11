package porter

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/stretchr/testify/require"
)

func TestNewDisplayInstallation(t *testing.T) {
	bun := bundle.Bundle{Name: "wordpress"}

	t.Run("install exists", func(t *testing.T) {
		install, err := claim.New("wordpress", claim.ActionInstall, bun, nil)
		require.NoError(t, err, "New claim failed")
		installResult, err := install.NewResult(claim.StatusSucceeded)
		require.NoError(t, err, "NewResult failed")

		upgrade, err := claim.New("wordpress", claim.ActionUpgrade, bun, nil)
		require.NoError(t, err, "New claim failed")
		upgradeResult, err := upgrade.NewResult(claim.StatusRunning)
		require.NoError(t, err, "NewResult failed")

		cp := claim.NewClaimStore(crud.NewMockStore(), nil, nil)
		err = cp.SaveClaim(install)
		require.NoError(t, err, "SaveClaim failed")
		err = cp.SaveResult(installResult)
		require.NoError(t, err, "SaveResult failed")
		err = cp.SaveClaim(upgrade)
		require.NoError(t, err, "SaveClaim failed")
		err = cp.SaveResult(upgradeResult)
		require.NoError(t, err, "SaveResult failed")

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
		dryRun, err := claim.New("wordpress", "dry-run", bun, nil)
		require.NoError(t, err, "New claim failed")
		dryRunResult, err := dryRun.NewResult(claim.StatusSucceeded)
		require.NoError(t, err, "NewResult failed")

		cp := claim.NewClaimStore(crud.NewMockStore(), nil, nil)
		err = cp.SaveClaim(dryRun)
		require.NoError(t, err, "SaveClaim failed")
		err = cp.SaveResult(dryRunResult)
		require.NoError(t, err, "SaveResult failed")

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
