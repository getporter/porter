package porter

import (
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

func TestPorter_printOutputsTable(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.CNAB = NewTestCNABProvider()

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
		Outputs: map[string]interface{}{
			"foo": "foo-value",
			"bar": "bar-value",
		},
	}

	err := p.InstanceStorage.Store(claim)
	require.NoError(t, err, "could not store claim")

	want := `-----------------------------------------
  Name  Type    Value (Path if sensitive)  
-----------------------------------------
  bar   string  bar-value                  
  foo   string  /path/to/foo               
`

	err = p.printOutputsTable(claim)
	require.NoError(t, err)

	got := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, want, got)
}
