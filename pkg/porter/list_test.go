package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
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

		i, err := cp.GetInstallation(context.Background(), "", "wordpress")
		require.NoError(t, err, "ReadInstallation failed")

		di := NewDisplayInstallation(i)

		require.Equal(t, di.Name, i.Name, "invalid installation name")
		require.Equal(t, di.Status.Created, i.Status.Created, "invalid created time")
		require.Equal(t, di.Status.Modified, i.Status.Modified, "invalid modified time")
		require.Equal(t, cnab.ActionUpgrade, di.Status.Action, "invalid last action")
		require.Equal(t, cnab.StatusRunning, di.Status.ResultStatus, "invalid last status")
	})

	t.Run("installation has not been installed", func(t *testing.T) {
		cp := claims.NewTestClaimProvider(t)
		defer cp.Teardown()

		i := cp.CreateInstallation(claims.NewInstallation("", "wordpress"))

		i, err := cp.GetInstallation(context.Background(), "", "wordpress")
		require.NoError(t, err, "GetInst failed")

		di := NewDisplayInstallation(i)

		require.Equal(t, di.Name, i.Name, "invalid installation name")
		require.Equal(t, i.Status.Created, di.Status.Created, "invalid created time")
		require.Equal(t, i.Status.Modified, di.Status.Modified, "invalid modified time")
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

func TestDisplayInstallation_ConvertToInstallation(t *testing.T) {
	cp := claims.NewTestClaimProvider(t)
	defer cp.Teardown()

	i := cp.CreateInstallation(claims.NewInstallation("", "wordpress"), func(i *claims.Installation) {
		i.Status.Action = cnab.ActionUpgrade
		i.Status.ResultStatus = cnab.StatusRunning
	})

	i, err := cp.GetInstallation(context.Background(), "", "wordpress")
	require.NoError(t, err, "ReadInstallation failed")

	di := NewDisplayInstallation(i)

	convertedInstallation, err := di.ConvertToInstallation()
	require.NoError(t, err, "failed to convert display installation to installation record")

	require.Equal(t, i.SchemaVersion, convertedInstallation.SchemaVersion, "invalid schema version")
	require.Equal(t, i.Name, convertedInstallation.Name, "invalid installation name")
	require.Equal(t, i.Namespace, convertedInstallation.Namespace, "invalid installation namespace")
	require.Equal(t, i.Uninstalled, convertedInstallation.Uninstalled, "invalid installation unstalled status")
	require.Equal(t, i.Bundle.Digest, convertedInstallation.Bundle.Digest, "invalid installation bundle")

	require.Equal(t, len(i.Labels), len(convertedInstallation.Labels))
	for key := range di.Labels {
		require.Equal(t, i.Labels[key], convertedInstallation.Labels[key], "invalid installation lables")
	}

	require.Equal(t, i.Custom, convertedInstallation.Custom, "invalid installation custom")

	require.Equal(t, convertedInstallation.CredentialSets, i.CredentialSets, "invalid credential set")
	require.Equal(t, convertedInstallation.ParameterSets, i.ParameterSets, "invalid parameter set")

	require.Equal(t, i.Parameters.String(), convertedInstallation.Parameters.String(), "invalid parameters name")
	require.Equal(t, len(i.Parameters.Parameters), len(convertedInstallation.Parameters.Parameters))

	parametersMap := make(map[string]secrets.Strategy, len(i.Parameters.Parameters))
	for _, param := range i.Parameters.Parameters {
		parametersMap[param.Name] = param
	}

	for _, param := range convertedInstallation.Parameters.Parameters {
		expected := parametersMap[param.Name]
		require.Equal(t, expected.Value, param.Value)
		expectedSource, err := expected.Source.MarshalJSON()
		require.NoError(t, err)
		source, err := param.Source.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, expectedSource, source)
	}

	require.Equal(t, i.Status.Created, convertedInstallation.Status.Created, "invalid created time")
	require.Equal(t, i.Status.Modified, convertedInstallation.Status.Modified, "invalid modified time")
	require.Equal(t, cnab.ActionUpgrade, convertedInstallation.Status.Action, "invalid last action")
	require.Equal(t, cnab.StatusRunning, convertedInstallation.Status.ResultStatus, "invalid last status")

}
