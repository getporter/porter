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

	// Create test claims
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
	c := p.TestClaims.CreateClaim("test", claim.ActionInstall, b, nil)
	c.Created = time.Date(2020, time.April, 18, 1, 2, 3, 4, time.UTC)
	err := p.Claims.SaveClaim(c)
	require.NoError(t, err, "SaveClaim failed")
	r := p.TestClaims.CreateResult(c, claim.StatusSucceeded)
	p.TestClaims.CreateOutput(c, r, "foo", []byte("foo-output"))
	p.TestClaims.CreateOutput(c, r, "bar", []byte("bar-output"))

	c = p.TestClaims.CreateClaim("test", claim.ActionUpgrade, b, nil)
	c.Created = time.Date(2020, time.April, 19, 1, 2, 3, 4, time.UTC)
	err = p.Claims.SaveClaim(c)
	require.NoError(t, err, "SaveClaim failed")
	r = p.TestClaims.CreateResult(c, claim.StatusRunning)

	err = p.ShowInstallation(opts)
	require.NoError(t, err, "ShowInstallation failed")

	wantOutput :=
		`Name: test
Created: 2020-04-18
Modified: 2020-04-19

Outputs:
----------------------------
  Name  Type    Value       
----------------------------
  bar   string  bar-output  
  foo   string  foo-output  

History:
----------------------------------
  Action   Timestamp   Status     
----------------------------------
  install  2020-04-18  succeeded  
  upgrade  2020-04-19  running    
`
	gotOutput := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, wantOutput, gotOutput)
}
