package porter

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/deislabs/cnab-go/claim"
	cnab "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/printer"
)

func TestPorter_ShowBundle(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	homeDir, err := p.TestConfig.GetHomeDir()
	require.NoError(t, err)

	p.TestConfig.TestContext.AddTestDirectory("testdata/outputs", filepath.Join(homeDir, "outputs"))

	opts := ShowOptions{
		Name:   "test-bundle",
		Format: printer.FormatTable,
	}
	d := cnab.NewTestDuffle(p.TestConfig)

	// Create test claim
	claim := claim.Claim{
		Name:     "test-bundle",
		Created:  time.Date(1983, time.April, 18, 1, 2, 3, 4, time.UTC),
		Modified: time.Date(1983, time.April, 18, 1, 2, 3, 4, time.UTC),
		Result: claim.Result{
			Action: "install",
			Status: "success",
		},
	}
	claimBytes, err := json.MarshalIndent(claim, "", "  ")
	require.NoError(t, err)
	d.ClaimStore.Store("test-bundle", claimBytes)

	err = p.ShowBundle(opts, d)
	require.NoError(t, err)

	wantOutput :=
		`Name: test-bundle
Created: 1983-04-18
Modified: 1983-04-18
Last Action: install
Last Status: success

Outputs:
-----------------------------------------------------
  Name  Type    Value (Path if sensitive)              
-----------------------------------------------------
  foo   string  /root/.porter/outputs/test-bundle/foo  
  bar   string  bar-value                              
`

	gotOutput := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, wantOutput, gotOutput)
}
