package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"github.com/stretchr/testify/assert"
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
		require.Equal(t, cnab.ActionUpgrade, di.Status.Action, "invalid last action")
		require.Equal(t, cnab.StatusRunning, di.Status.ResultStatus, "invalid last status")
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
		require.Empty(t, di.Status.Action, "invalid last action")
		require.Empty(t, di.Status.ResultStatus, "invalid last status")
	})
}

func TestPorter_ListInstallations(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestClaims.CreateInstallation(claims.NewInstallation("", "shared-mysql"))
	p.TestClaims.CreateInstallation(claims.NewInstallation("dev", "carolyn-wordpress"))
	p.TestClaims.CreateInstallation(claims.NewInstallation("dev", "vaughn-wordpress"))
	p.TestClaims.CreateInstallation(claims.NewInstallation("test", "staging-wordpress"))
	p.TestClaims.CreateInstallation(claims.NewInstallation("test", "iat-wordpress"))
	p.TestClaims.CreateInstallation(claims.NewInstallation("test", "shared-mysql"))

	t.Run("all-namespaces", func(t *testing.T) {
		opts := ListOptions{AllNamespaces: true}
		results, err := p.ListInstallations(context.Background(), opts)
		require.NoError(t, err)
		assert.Len(t, results, 6)
	})

	t.Run("local namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: "dev"}
		results, err := p.ListInstallations(context.Background(), opts)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		opts = ListOptions{Namespace: "test"}
		results, err = p.ListInstallations(context.Background(), opts)
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("global namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: ""}
		results, err := p.ListInstallations(context.Background(), opts)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}
