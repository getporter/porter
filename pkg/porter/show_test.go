package porter

import (
	"fmt"
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
	c1 := p.TestClaims.CreateClaim("test", claim.ActionInstall, b, nil)
	c1.Created = time.Date(2020, time.April, 18, 1, 2, 3, 4, time.UTC)
	err := p.Claims.SaveClaim(c1)
	require.NoError(t, err, "SaveClaim failed")
	r := p.TestClaims.CreateResult(c1, claim.StatusSucceeded)
	p.TestClaims.CreateOutput(c1, r, "foo", []byte("foo-output"))
	p.TestClaims.CreateOutput(c1, r, "bar", []byte("bar-output"))

	c2 := p.TestClaims.CreateClaim("test", claim.ActionUpgrade, b, nil)
	c2.Created = time.Date(2020, time.April, 19, 1, 2, 3, 4, time.UTC)
	err = p.Claims.SaveClaim(c2)
	require.NoError(t, err, "SaveClaim failed")
	r = p.TestClaims.CreateResult(c2, claim.StatusRunning)

	err = p.ShowInstallation(opts)
	require.NoError(t, err, "ShowInstallation failed")

	wantOutput := fmt.Sprintf(
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
------------------------------------------------------------------------
  Run ID                      Action   Timestamp   Status     Has Logs  
------------------------------------------------------------------------
  %s  install  2020-04-18  succeeded  false     
  %s  upgrade  2020-04-19  running    false     
`, c1.ID, c2.ID)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, wantOutput, gotOutput)
}
