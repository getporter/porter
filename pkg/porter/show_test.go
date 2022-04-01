package porter

import (
	"context"
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

	ref := "getporter/wordpress:v0.1.0"
	testcases := []struct {
		name       string
		ref        string
		format     printer.Format
		outputFile string
	}{
		{name: "plain", ref: ref, format: printer.FormatPlaintext, outputFile: "testdata/show/expected-output.txt"},
		{name: "no reference, plain", format: printer.FormatPlaintext, outputFile: "testdata/show/no-reference-expected-output.txt"},
		{name: "json", ref: ref, format: printer.FormatJson, outputFile: "testdata/show/expected-output.json"},
		{name: "yaml", ref: ref, format: printer.FormatYaml, outputFile: "testdata/show/expected-output.yaml"},
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
				if tc.ref != "" {
					i.TrackBundle(cnab.MustParseOCIReference(tc.ref))
				}
				i.Labels = map[string]string{
					"io.cnab/app":        "wordpress",
					"io.cnab/appVersion": "v1.2.3",
				}
				i.Parameters = map[string]interface{}{"logLevel": 3}
				i.ParameterSets = []string{"dev-env"}
			})
			run := p.TestClaims.CreateRun(i.NewRun(cnab.ActionUpgrade), p.TestClaims.SetMutableRunValues, func(r *claims.Run) {
				r.Bundle = b
				r.BundleReference = tc.ref
				r.BundleDigest = "sha256:88d68ef0bdb9cedc6da3a8e341a33e5d2f8bb19d0cf7ec3f1060d3f9eb73cae9"
				r.Parameters = map[string]interface{}{"token": "top-secret", "logLevel": 3}
			})
			result := p.TestClaims.CreateResult(run.NewResult(cnab.StatusSucceeded), p.TestClaims.SetMutableResultValues)
			i.ApplyResult(run, result)
			i.Status.Installed = &now
			require.NoError(t, p.TestClaims.UpdateInstallation(i))

			err := p.ShowInstallation(context.Background(), opts)
			require.NoError(t, err, "ShowInstallation failed")
			p.CompareGoldenFile(tc.outputFile, p.TestConfig.TestContext.GetOutput())
		})
	}
}
