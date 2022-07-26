package porter

import (
	"context"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2020, time.April, 18, 1, 2, 3, 4, time.UTC)

func TestPorter_ListInstallationRuns(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	installationName1 := "shared-mysql"
	run1 := storage.NewRun("", installationName1)
	run1.NewResult("running")
	run1.NewResult("succeeded")

	p.TestInstallations.CreateInstallation(storage.NewInstallation("", installationName1), p.TestInstallations.SetMutableInstallationValues)
	p.TestInstallations.CreateRun(run1)

	installationName2 := "shared-k8s"

	run2 := storage.NewRun("dev", installationName2)
	run2.NewResult("running")

	run3 := storage.NewRun("dev", installationName2)
	run3.NewResult("running")

	p.TestInstallations.CreateInstallation(storage.NewInstallation("dev", installationName2), p.TestInstallations.SetMutableInstallationValues)
	p.TestInstallations.CreateRun(run2)
	p.TestInstallations.CreateRun(run3)

	t.Run("global namespace", func(t *testing.T) {
		opts := RunListOptions{installationOptions: installationOptions{
			Namespace: "",
			Name:      installationName1,
		}}
		results, err := p.ListInstallationRuns(context.Background(), opts)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("specified namespace", func(t *testing.T) {
		opts := RunListOptions{installationOptions: installationOptions{
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()
			ctx := context.Background()

			installation := p.TestInstallations.CreateInstallation(storage.NewInstallation("staging", "shared-k8s"), p.TestInstallations.SetMutableInstallationValues)

			installRun := p.TestInstallations.CreateRun(installation.NewRun(cnab.ActionInstall), p.TestInstallations.SetMutableRunValues)
			uninstallRun := p.TestInstallations.CreateRun(installation.NewRun(cnab.ActionUninstall), p.TestInstallations.SetMutableRunValues)
			result := p.TestInstallations.CreateResult(installRun.NewResult(cnab.StatusSucceeded), p.TestInstallations.SetMutableResultValues)
			result2 := p.TestInstallations.CreateResult(uninstallRun.NewResult(cnab.StatusSucceeded), p.TestInstallations.SetMutableResultValues)

			installation.ApplyResult(installRun, result)
			installation.ApplyResult(uninstallRun, result2)
			installation.Status.Installed = &now

			require.NoError(t, p.TestInstallations.UpdateInstallation(ctx, installation))

			opts := RunListOptions{installationOptions: installationOptions{
				Namespace: "staging",
				Name:      "shared-k8s",
			}, PrintOptions: printer.PrintOptions{Format: tc.format},
			}

			err := p.PrintInstallationRuns(context.Background(), opts)
			require.NoError(t, err)

			p.CompareGoldenFile(tc.outputFile, p.TestConfig.TestContext.GetOutput())
		})

	}
}
