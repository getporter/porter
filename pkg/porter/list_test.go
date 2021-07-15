package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"github.com/stretchr/testify/require"
)

func TestNewDisplayInstallation(t *testing.T) {
	t.Run("installation has been installed", func(t *testing.T) {
		cp := claims.NewTestClaimProvider(t)
		defer cp.Teardown()

		i := cp.CreateInstallation(claims.NewInstallation("", "wordpress"), func(i *claims.Installation) {
			i.Status.Action = cnab.ActionUpgrade
			i.Status.ResultStatus = cnab.StatusRunning
		})

		i, err := cp.GetInstallation("", "wordpress")
		require.NoError(t, err, "ReadInstallation failed")

		di := NewDisplayInstallation(i, nil)

		require.Equal(t, di.Name, i.Name, "invalid installation name")
		require.Equal(t, di.Created, i.Created, "invalid created time")
		require.Equal(t, di.Modified, i.Modified, "invalid modified time")
		require.Equal(t, cnab.ActionUpgrade, di.Action, "invalid last action")
		require.Equal(t, cnab.StatusRunning, di.Status, "invalid last status")
	})

	t.Run("installation has not been installed", func(t *testing.T) {
		cp := claims.NewTestClaimProvider(t)
		defer cp.Teardown()

		i := cp.CreateInstallation(claims.NewInstallation("", "wordpress"))

		i, err := cp.GetInstallation("", "wordpress")
		require.NoError(t, err, "GetInst failed")

		di := NewDisplayInstallation(i, nil)

		require.Equal(t, di.Name, i.Name, "invalid installation name")
		require.Equal(t, i.Created, di.Created, "invalid created time")
		require.Equal(t, i.Modified, di.Modified, "invalid modified time")
		require.Empty(t, di.Action, "invalid last action")
		require.Empty(t, di.Status, "invalid last status")
	})
}
