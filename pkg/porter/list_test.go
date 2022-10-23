package porter

import (
	"context"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDisplayInstallation(t *testing.T) {
	t.Run("installation has been installed", func(t *testing.T) {
		cp := storage.NewTestInstallationProvider(t)
		defer cp.Close()

		i := cp.CreateInstallation(storage.NewInstallation("", "wordpress"), func(i *storage.Installation) {
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
		cp := storage.NewTestInstallationProvider(t)
		defer cp.Close()

		i := cp.CreateInstallation(storage.NewInstallation("", "wordpress"))

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
	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	i1 := storage.NewInstallation("", "shared-mysql")
	i1.Parameters.Parameters = []secrets.Strategy{ // Define a parameter that is stored in a secret, list should not retrieve it
		{Name: "password", Source: secrets.Source{Key: "secret", Value: "mypassword"}},
	}
	i1.Status.RunID = "10" // Add a run but don't populate the data for it, list should not retrieve it

	p.TestInstallations.CreateInstallation(i1)
	p.TestInstallations.CreateInstallation(storage.NewInstallation("dev", "carolyn-wordpress"))
	p.TestInstallations.CreateInstallation(storage.NewInstallation("dev", "vaughn-wordpress"))
	p.TestInstallations.CreateInstallation(storage.NewInstallation("test", "staging-wordpress"))
	p.TestInstallations.CreateInstallation(storage.NewInstallation("test", "iat-wordpress"))
	p.TestInstallations.CreateInstallation(storage.NewInstallation("test", "shared-mysql"))

	t.Run("all-namespaces", func(t *testing.T) {
		opts := ListOptions{AllNamespaces: true}
		results, err := p.ListInstallations(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 6)

		// Check that porter didn't go off and retrieve extra data for each installation
		for _, r := range results {
			assert.Empty(t, r.ResolvedParameters, "ListInstallations should not resolve secrets used by the installations")
		}
	})

	t.Run("local namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: "dev"}
		results, err := p.ListInstallations(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		opts = ListOptions{Namespace: "test"}
		results, err = p.ListInstallations(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("global namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: ""}
		results, err := p.ListInstallations(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestDisplayInstallation_ConvertToInstallation(t *testing.T) {
	cp := storage.NewTestInstallationProvider(t)
	defer cp.Close()

	i := cp.CreateInstallation(storage.NewInstallation("", "wordpress"), func(i *storage.Installation) {
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

func TestPorter_PrintInstallations(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		format     printer.Format
		outputFile string
	}{
		{name: "plain", format: printer.FormatPlaintext, outputFile: "testdata/list/expected-output.txt"},
		{name: "no reference, plain", format: printer.FormatPlaintext, outputFile: "testdata/list/no-reference-expected-output.txt"},
		{name: "json", format: printer.FormatJson, outputFile: "testdata/list/expected-output.json"},
		{name: "yaml", format: printer.FormatYaml, outputFile: "testdata/list/expected-output.yaml"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			opts := ListOptions{
				Namespace: "dev",
				Name:      "mywordpress",
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
			}

			p.TestInstallations.CreateInstallation(storage.NewInstallation("dev", "mywordpress"), p.TestInstallations.SetMutableInstallationValues, func(i *storage.Installation) {
				i.Status.BundleVersion = "v1.2.3"
				i.Status.ResultStatus = cnab.StatusSucceeded
			})

			ctx := context.Background()

			err := p.PrintInstallations(ctx, opts)
			require.NoError(t, err, "PrintInstallation failed")
			p.CompareGoldenFile(tc.outputFile, p.TestConfig.TestContext.GetOutput())
		})
	}
}

func TestPorter_getDisplayInstallationState(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	installation := p.TestInstallations.CreateInstallation(storage.NewInstallation("dev", "mywordpress"), p.TestInstallations.SetMutableInstallationValues)
	displayInstallationState := getDisplayInstallationState(installation)
	require.Equal(t, StateDefined, displayInstallationState)

	run := p.TestInstallations.CreateRun(installation.NewRun(cnab.ActionInstall), p.TestInstallations.SetMutableRunValues)
	result := p.TestInstallations.CreateResult(run.NewResult(cnab.StatusSucceeded), p.TestInstallations.SetMutableResultValues)
	installation.ApplyResult(run, result)
	installTime := now.Add(-time.Second * 5)
	installation.Status.Installed = &installTime
	displayInstallationState = getDisplayInstallationState(installation)
	require.Equal(t, StateInstalled, displayInstallationState)

	run = p.TestInstallations.CreateRun(installation.NewRun(cnab.ActionUninstall), p.TestInstallations.SetMutableRunValues)
	result = p.TestInstallations.CreateResult(run.NewResult(cnab.StatusSucceeded), p.TestInstallations.SetMutableResultValues)
	installation.ApplyResult(run, result)
	installation.Status.Uninstalled = &now
	displayInstallationState = getDisplayInstallationState(installation)
	require.Equal(t, StateUninstalled, displayInstallationState)
}

func TestPorter_getDisplayInstallationStatus(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	installation := p.TestInstallations.CreateInstallation(storage.NewInstallation("dev", "mywordpress"), p.TestInstallations.SetMutableInstallationValues)
	run := p.TestInstallations.CreateRun(installation.NewRun(cnab.ActionInstall), p.TestInstallations.SetMutableRunValues)
	result := p.TestInstallations.CreateResult(run.NewResult(cnab.StatusSucceeded), p.TestInstallations.SetMutableResultValues)
	installation.ApplyResult(run, result)
	displayInstallationStatus := getDisplayInstallationStatus(installation)
	require.Equal(t, cnab.StatusSucceeded, displayInstallationStatus)

	result = p.TestInstallations.CreateResult(run.NewResult(cnab.StatusFailed), p.TestInstallations.SetMutableResultValues)
	installation.ApplyResult(run, result)
	displayInstallationStatus = getDisplayInstallationStatus(installation)
	require.Equal(t, cnab.StatusFailed, displayInstallationStatus)

	run = p.TestInstallations.CreateRun(installation.NewRun(cnab.ActionInstall), p.TestInstallations.SetMutableRunValues)
	result = p.TestInstallations.CreateResult(run.NewResult(cnab.StatusRunning), p.TestInstallations.SetMutableResultValues)
	installation.ApplyResult(run, result)
	displayInstallationStatus = getDisplayInstallationStatus(installation)
	require.Equal(t, StatusInstalling, displayInstallationStatus)

	run = p.TestInstallations.CreateRun(installation.NewRun(cnab.ActionUninstall), p.TestInstallations.SetMutableRunValues)
	result = p.TestInstallations.CreateResult(run.NewResult(cnab.StatusRunning), p.TestInstallations.SetMutableResultValues)
	installation.ApplyResult(run, result)
	displayInstallationStatus = getDisplayInstallationStatus(installation)
	require.Equal(t, StatusUninstalling, displayInstallationStatus)

	run = p.TestInstallations.CreateRun(installation.NewRun(cnab.ActionUpgrade), p.TestInstallations.SetMutableRunValues)
	result = p.TestInstallations.CreateResult(run.NewResult(cnab.StatusRunning), p.TestInstallations.SetMutableResultValues)
	installation.ApplyResult(run, result)
	displayInstallationStatus = getDisplayInstallationStatus(installation)
	require.Equal(t, StatusUpgrading, displayInstallationStatus)

	run = p.TestInstallations.CreateRun(installation.NewRun("customaction"), p.TestInstallations.SetMutableRunValues)
	result = p.TestInstallations.CreateResult(run.NewResult(cnab.StatusRunning), p.TestInstallations.SetMutableResultValues)
	installation.ApplyResult(run, result)
	installation.Status.Action = "customaction"
	displayInstallationStatus = getDisplayInstallationStatus(installation)
	require.Equal(t, "running customaction", displayInstallationStatus)
}
