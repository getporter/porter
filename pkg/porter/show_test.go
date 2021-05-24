package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

func TestPorter_ShowBundle(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		format     printer.Format
		outputFile string
	}{
		{name: "plain", format: printer.FormatPlaintext, outputFile: "testdata/show/expected-output.txt"},
		{name: "json", format: printer.FormatJson, outputFile: "testdata/show/expected-output.json"},
		{name: "yaml", format: printer.FormatYaml, outputFile: "testdata/show/expected-output.yaml"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)

			opts := ShowOptions{
				sharedOptions: sharedOptions{
					Name: "test",
				},
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
			}

			// Create test claims
			writeOnly := true
			b := bundle.Bundle{
				Name:    "porter-hello",
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
			c1 := p.TestClaims.CreateClaim("test", claim.ActionInstall, b, map[string]interface{}{"token": "top-secret", "logLevel": 5})
			r := p.TestClaims.CreateResult(c1, claim.StatusSucceeded)
			p.TestClaims.CreateOutput(c1, r, "foo", []byte("foo-output"))
			p.TestClaims.CreateOutput(c1, r, "bar", []byte("bar-output"))

			c2 := p.TestClaims.CreateClaim("test", claim.ActionUpgrade, b, map[string]interface{}{"token": "top-secret", "logLevel": 3})
			r = p.TestClaims.CreateResult(c2, claim.StatusRunning)

			err := p.ShowInstallation(opts)
			require.NoError(t, err, "ShowInstallation failed")
			p.CompareGoldenFile(tc.outputFile, p.TestConfig.TestContext.GetOutput())
		})
	}
}
