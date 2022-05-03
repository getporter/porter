package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
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
			defer p.Close()

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
					"secretString": {
						Definition: "secretString",
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

			bun := cnab.ExtendedBundle{b}
			i := p.TestClaims.CreateInstallation(storage.NewInstallation("dev", "mywordpress"), p.TestClaims.SetMutableInstallationValues, func(i *storage.Installation) {
				if tc.ref != "" {
					i.TrackBundle(cnab.MustParseOCIReference(tc.ref))
				}
				i.Labels = map[string]string{
					"io.cnab/app":        "wordpress",
					"io.cnab/appVersion": "v1.2.3",
				}
				params := []secrets.Strategy{
					{Name: "logLevel", Source: secrets.Source{Value: "3"}, Value: "3"},
					secrets.Strategy{Name: "secretString", Source: secrets.Source{Key: "secretString", Value: "foo"}, Value: "foo"},
				}
				i.Parameters = i.NewInternalParameterSet(params...)

				i.ParameterSets = []string{"dev-env"}

				i.Parameters.Parameters = p.SanitizeParameters(i.Parameters.Parameters, i.ID, bun)
			})

			run := p.TestClaims.CreateRun(i.NewRun(cnab.ActionUpgrade), p.TestClaims.SetMutableRunValues, func(r *storage.Run) {
				r.Bundle = b
				r.BundleReference = tc.ref
				r.BundleDigest = "sha256:88d68ef0bdb9cedc6da3a8e341a33e5d2f8bb19d0cf7ec3f1060d3f9eb73cae9"

				r.ParameterOverrides = i.NewInternalParameterSet(
					storage.ValueStrategy("logLevel", "3"),
					storage.ValueStrategy("secretString", "foo"),
				)

				r.Parameters = i.NewInternalParameterSet(
					[]secrets.Strategy{
						storage.ValueStrategy("logLevel", "3"),
						storage.ValueStrategy("token", "top-secret"),
						storage.ValueStrategy("secretString", "foo"),
					}...)

				r.ParameterSets = []string{"dev-env"}
				r.ParameterOverrides.Parameters = p.SanitizeParameters(r.ParameterOverrides.Parameters, r.ID, bun)
				r.Parameters.Parameters = p.SanitizeParameters(r.Parameters.Parameters, r.ID, bun)
			})

			i.Parameters.Parameters = run.ParameterOverrides.Parameters
			err := p.TestClaims.UpsertInstallation(context.Background(), i)
			require.NoError(t, err)

			result := p.TestClaims.CreateResult(run.NewResult(cnab.StatusSucceeded), p.TestClaims.SetMutableResultValues)
			i.ApplyResult(run, result)
			i.Status.Installed = &now
			ctx := context.Background()
			require.NoError(t, p.TestClaims.UpdateInstallation(ctx, i))

			err = p.ShowInstallation(ctx, opts)
			require.NoError(t, err, "ShowInstallation failed")
			p.CompareGoldenFile(tc.outputFile, p.TestConfig.TestContext.GetOutput())
		})
	}
}
