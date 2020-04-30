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
	b := bundle.Bundle{
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
	}
	c, err := claim.New("test", claim.ActionInstall, b, nil)
	require.NoError(t, err, "NewClaim failed")
	c.Created = time.Date(1983, time.April, 18, 1, 2, 3, 4, time.UTC)
	err = p.Claims.SaveClaim(c)
	require.NoError(t, err, "SaveClaim failed")

	r, err := c.NewResult(claim.StatusSucceeded)
	require.NoError(t, err, "NewResult failed")
	err = p.Claims.SaveResult(r)
	require.NoError(t, err, "SaveResult failed")

	foo := claim.NewOutput(c, r, "foo", []byte("foo-output"))
	err = p.Claims.SaveOutput(foo)
	require.NoError(t, err, "SaveOutput failed")

	bar := claim.NewOutput(c, r, "bar", []byte("bar-output"))
	err = p.Claims.SaveOutput(bar)
	require.NoError(t, err, "SaveOutput failed")

	err = p.ShowInstallations(opts)
	require.NoError(t, err, "ShowInstallations failed")

	wantOutput :=
		`Name: test
Created: 1983-04-18
Modified: 1983-04-18
Last Action: install
Last Status: succeeded

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
