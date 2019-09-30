package porter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/porter/pkg/printer"
)

func TestPorter_ShowBundle(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.CNAB = NewTestCNABProvider()

	opts := ShowOptions{
		sharedOptions: sharedOptions{
			Name: "test-bundle",
		},
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatTable,
		},
	}

	// Create test claim
	writeOnly := true
	claim := claim.Claim{
		Name: "test-bundle",
		Bundle: &bundle.Bundle{
			Definitions: definition.Definitions{
				"foo": &definition.Schema{
					Type:      "string",
					WriteOnly: &writeOnly,
				},
				"bar": &definition.Schema{
					Type: "string",
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
			},
		},
		Created:  time.Date(1983, time.April, 18, 1, 2, 3, 4, time.UTC),
		Modified: time.Date(1983, time.April, 18, 1, 2, 3, 4, time.UTC),
		Result: claim.Result{
			Action: "install",
			Status: "success",
		},
		Outputs: map[string]interface{}{
			"foo": "foo-output",
			"bar": "bar-output",
		},
	}
	if testy, ok := p.CNAB.(*TestCNABProvider); ok {
		testy.CreateClaim(&claim)
	} else {
		t.Fatal("expected p.CNAB to be of type *TestCNABProvider")
	}

	err := p.ShowInstances(opts)
	require.NoError(t, err)

	wantOutput :=
		`Name: test-bundle
Created: 1983-04-18
Modified: 1983-04-18
Last Action: install
Last Status: success

Outputs:
-----------------------------------------
  Name  Type    Value (Path if sensitive)  
-----------------------------------------
  bar   string  bar-output                 
  foo   string  /path/to/foo               
`

	gotOutput := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, wantOutput, gotOutput)
}
