package porter

import (
	"context"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2020, time.April, 18, 1, 2, 3, 4, time.UTC)

func TestPorter_ListInstallationRuns(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	installationName1 := "shared-mysql"
	run1 := claims.NewRun("", installationName1)
	run1.NewResult("running")
	run1.NewResult("succeeded")

	p.TestClaims.CreateInstallation(claims.NewInstallation("", installationName1), p.TestClaims.SetMutableInstallationValues)
	p.TestClaims.CreateRun(run1)

	installationName2 := "shared-k8s"

	run2 := claims.NewRun("dev", installationName2)
	run2.NewResult("running")

	run3 := claims.NewRun("dev", installationName2)
	run3.NewResult("running")

	p.TestClaims.CreateInstallation(claims.NewInstallation("dev", installationName2), p.TestClaims.SetMutableInstallationValues)
	p.TestClaims.CreateRun(run2)
	p.TestClaims.CreateRun(run3)

	t.Run("global namespace", func(t *testing.T) {
		opts := RunListOptions{sharedOptions: sharedOptions{
			Namespace: "",
			Name:      installationName1,
		}}
		results, err := p.ListInstallationRuns(context.Background(), opts)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("specified namespace", func(t *testing.T) {
		opts := RunListOptions{sharedOptions: sharedOptions{
			Namespace: "dev",
			Name:      installationName2,
		}}
		results, err := p.ListInstallationRuns(context.Background(), opts)
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})
}

func TestPorter_PrintInstallationRunsOutput(t *testing.T) {
	outputTestcases := []struct {
		name       string
		format     printer.Format
		outputFile string
	}{
		{name: "yaml", format: printer.FormatYaml, outputFile: "testdata/runs/expected-output.yaml"},
		{name: "json", format: printer.FormatJson, outputFile: "testdata/runs/expected-output.json"},
		{name: "plaintext", format: printer.FormatPlaintext, outputFile: "testdata/runs/expected-output.txt"},
	}

	for _, tc := range outputTestcases {
		p := NewTestPorter(t)
		defer p.Close()
		ctx := context.Background()

		installation := p.TestClaims.CreateInstallation(claims.NewInstallation("staging", "shared-k8s"), p.TestClaims.SetMutableInstallationValues)

		installRun := p.TestClaims.CreateRun(installation.NewRun(cnab.ActionInstall), p.TestClaims.SetMutableRunValues)
		uninstallRun := p.TestClaims.CreateRun(installation.NewRun(cnab.ActionUninstall), p.TestClaims.SetMutableRunValues)
		result := p.TestClaims.CreateResult(installRun.NewResult(cnab.StatusSucceeded), p.TestClaims.SetMutableResultValues)
		result2 := p.TestClaims.CreateResult(uninstallRun.NewResult(cnab.StatusSucceeded), p.TestClaims.SetMutableResultValues)

		installation.ApplyResult(installRun, result)
		installation.ApplyResult(uninstallRun, result2)
		installation.Status.Installed = &now

		require.NoError(t, p.TestClaims.UpdateInstallation(ctx, installation))

		opts := RunListOptions{sharedOptions: sharedOptions{
			Namespace: "staging",
			Name:      "shared-k8s",
		}, PrintOptions: printer.PrintOptions{Format: tc.format},
		}

		err := p.PrintInstallationRuns(context.Background(), opts)
		require.NoError(t, err)

		p.CompareGoldenFile(tc.outputFile, p.TestConfig.TestContext.GetOutput())
	}
}
