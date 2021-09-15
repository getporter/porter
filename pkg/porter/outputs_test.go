package porter

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/require"
)

func TestPorter_printOutputsTable(t *testing.T) {
	t.Parallel()

	b, err := ioutil.ReadFile("testdata/show/object-parameter-value.json")
	require.NoError(t, err)
	var objVal map[string]interface{}
	err = json.Unmarshal(b, &objVal)
	require.NoError(t, err)

	p := NewTestPorter(t)
	defer p.Teardown()

	want := `---------------------------------------------------------------------------------
  Name     Type    Value                                                         
---------------------------------------------------------------------------------
  bar      string  ******                                                        
  foo      string  /path/to/foo                                                  
  object   object  {"a":{"b":1,"c":2},"d":"yay"}                                 
  longfoo  string  DFo6Wc2jDhmA7Yt4PbHyh8RO4vVG7leOzK412gf2TXNPJhuCUs1rB29nk...  
`

	outputs := DisplayValues{
		{Name: "bar", Type: "string", Value: "bar-value", Sensitive: true},
		{Name: "foo", Type: "string", Value: "/path/to/foo"},
		{Name: "object", Type: "object", Value: objVal},
		{Name: "longfoo", Type: "string", Value: "DFo6Wc2jDhmA7Yt4PbHyh8RO4vVG7leOzK412gf2TXNPJhuCUs1rB29nkJJd4ICimZGpyWpMGalSvDxf"},
	}
	err = p.printDisplayValuesTable(outputs)
	require.NoError(t, err)

	got := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, want, got)
}

func TestPorter_PrintBundleOutputs(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name           string
		format         printer.Format
		expectedOutput string
	}{
		{name: "text", format: printer.FormatPlaintext, expectedOutput: "testdata/outputs/show-expected-output.txt"},
		{name: "json", format: printer.FormatJson, expectedOutput: "testdata/outputs/show-expected-output.json"},
		{name: "yaml", format: printer.FormatYaml, expectedOutput: "testdata/outputs/show-expected-output.yaml"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Teardown()

			// Create test claim
			writeOnly := true
			b := bundle.Bundle{
				Definitions: definition.Definitions{
					"foo": &definition.Schema{
						Type:      "string",
						WriteOnly: &writeOnly,
					},
					"bar": &definition.Schema{
						Type: "string",
					},
					"longfoo": &definition.Schema{
						Type: "string",
					},
					"porter-state": &definition.Schema{
						Type:    "string",
						Comment: "porter-internal", // This output should be hidden because it's internal
					},
				},
				Outputs: map[string]bundle.Output{
					"foo": {
						Definition: "foo",
						Path:       "/path/to/foo",
					},
					"bar": {
						Definition: "bar",
					},
					"longfoo": {
						Definition: "longfoo",
					},
					"porter-state": {
						Definition: "porter-state",
						Path:       "/cnab/app/outputs/porter-state.tgz",
					},
				},
			}

			i := p.TestClaims.CreateInstallation(claims.NewInstallation("", "test"))
			c := p.TestClaims.CreateRun(i.NewRun(cnab.ActionInstall), func(r *claims.Run) { r.Bundle = b })
			r := p.TestClaims.CreateResult(c.NewResult(cnab.StatusSucceeded))
			p.TestClaims.CreateOutput(r.NewOutput("foo", []byte("foo-output")))
			p.TestClaims.CreateOutput(r.NewOutput("bar", []byte("bar-output")))
			p.TestClaims.CreateOutput(r.NewOutput("longfoo", []byte("DFo6Wc2jDhmA7Yt4PbHyh8RO4vVG7leOzK412gf2TXNPJhuCUs1rB29nkJJd4ICimZGpyWpMGalSvDxf")))
			p.TestClaims.CreateOutput(r.NewOutput("porter-state", []byte("porter-state.tgz contents")))

			opts := OutputListOptions{
				sharedOptions: sharedOptions{
					Name: "test",
				},
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
			}
			err := p.PrintBundleOutputs(opts)
			require.NoError(t, err, "could not print bundle outputs")

			p.CompareGoldenFile(tc.expectedOutput, p.TestConfig.TestContext.GetOutput())
		})
	}
}
