package porter

import (
	"testing"
	"time"

	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

func TestPorter_ShowBundle(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	opts := ShowOptions{
		sharedOptions: sharedOptions{
			Name: "test",
		},
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatTable,
		},
	}

	// Create test claim
	writeOnly := true
	claim := claim.Claim{
		Installation: "test",
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

	err := p.Claims.Save(claim)
	require.NoError(t, err, "could not store claim")

	err = p.ShowInstances(opts)
	require.NoError(t, err, "ShowInstances failed")

	wantOutput :=
		`Name: test
Created: 1983-04-18
Modified: 1983-04-18
Last Action: install
Last Status: success

Outputs:
----------------------------
  Name  Type    Value       
----------------------------
  bar   string  bar-output  
  foo   string  foo-output  
`
	gotOutput := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, wantOutput, gotOutput)
}
