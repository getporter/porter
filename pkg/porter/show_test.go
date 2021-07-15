package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/require"
)

func TestPorter_ShowBundle(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		repo       string
		format     printer.Format
		outputFile string
	}{
		{name: "plain", repo: "getporter/wordpress", format: printer.FormatPlaintext, outputFile: "testdata/show/expected-output.txt"},
		{name: "no reference, plain", repo: "", format: printer.FormatPlaintext, outputFile: "testdata/show/no-reference-expected-output.txt"},
		{name: "json", repo: "getporter/wordpress", format: printer.FormatJson, outputFile: "testdata/show/expected-output.json"},
		{name: "yaml", repo: "getporter/wordpress", format: printer.FormatYaml, outputFile: "testdata/show/expected-output.yaml"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Teardown()

			opts := ShowOptions{
				sharedOptions: sharedOptions{
					Namespace: "dev",
					Name:      "mywordpress",
				},
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
			}

			// Create test claims
			writeOnly := true
			b := bundle.Bundle{
				Name:    "wordpress",
				Version: "0.1.0",
				Definitions: definition.Definitions{
					"secretString": &definition.Schema{
						Type:      "string",
						WriteOnly: &writeOnly,
					},
					"bar": &definition.Schema{
						Type: "string",
					},
					"logLevel": &definition.Schema{
						Type: "integer",
					},
				},
				Parameters: map[string]bundle.Parameter{
					"logLevel": {
						Definition: "logLevel",
					},
					"token": {
						Definition: "foo",
					},
				},
				Outputs: map[string]bundle.Output{
					"foo": {
						Definition: "secretString",
						Path:       "/path/to/foo",
					},
					"bar": {
						Definition: "bar",
					},
				},
			}
			i := p.TestClaims.CreateInstallation(claims.NewInstallation("dev", "mywordpress"), p.TestClaims.SetMutableInstallationValues, func(i *claims.Installation) {
				i.BundleVersion = "0.1.0"
				i.BundleRepository = tc.repo
				i.BundleDigest = "sha256:88d68ef0bdb9cedc6da3a8e341a33e5d2f8bb19d0cf7ec3f1060d3f9eb73cae9"
				i.Labels = map[string]string{
					"io.cnab/app":        "wordpress",
					"io.cnab/appVersion": "v1.2.3",
				}
				i.Parameters = map[string]interface{}{"token": "top-secret", "logLevel": 3}
			})
			r := p.TestClaims.CreateRun(i.NewRun(cnab.ActionUpgrade), p.TestClaims.SetMutableRunValues, func(r *claims.Run) {
				r.Bundle = b
				if tc.repo != "" {
					r.BundleReference = tc.repo + ":0.1.0"
				}

			})
			i.Status.RunID = r.ID
			i.Status.Action = r.Action
			i.Status.ResultStatus = cnab.StatusSucceeded
			i.Status.InstallationCompleted = true
			require.NoError(t, p.TestClaims.UpdateInstallation(i))

			err := p.ShowInstallation(opts)
			require.NoError(t, err, "ShowInstallation failed")
			p.CompareGoldenFile(tc.outputFile, p.TestConfig.TestContext.GetOutput())
		})
	}
}
