package porter

import (
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

func TestPorter_fetchBundleOutputs_Error(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.CNAB = NewTestCNABProvider()

	_, err := p.fetchBundleOutputs("bad-outputs-bundle")
	require.EqualError(t, err,
		"could not read bundle instance file for bad-outputs-bundle: open bad-outputs-bundle: file does not exist")
}

func TestPorter_fetchBundleOutputs(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.CNAB = NewTestCNABProvider()

	// Create test claim
	claim := claim.Claim{
		Name: "test-bundle",
		Outputs: map[string]interface{}{
			"foo": "foo-value",
			"bar": "bar-value",
		},
	}
	if testy, ok := p.CNAB.(*TestCNABProvider); ok {
		testy.CreateClaim(&claim)
	} else {
		t.Fatal("expected p.CNAB to be of type *TestCNABProvider")
	}

	got, err := p.fetchBundleOutputs("test-bundle")
	require.NoError(t, err)

	want := map[string]interface{}{
		"foo": "foo-value",
		"bar": "bar-value",
	}

	require.Equal(t, want, got)
}

func TestPorter_printOutputsTable(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.CNAB = NewTestCNABProvider()

	outputs := map[string]interface{}{
		"foo": "foo-value",
		"bar": "bar-value",
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
		Outputs: map[string]interface{}{
			"foo": "foo-value",
			"bar": "bar-value",
		},
	}
	if testy, ok := p.CNAB.(*TestCNABProvider); ok {
		testy.CreateClaim(&claim)
	} else {
		t.Fatal("expected p.CNAB to be of type *TestCNABProvider")
	}

	want := `-----------------------------------------
  Name  Type    Value (Path if sensitive)  
-----------------------------------------
  bar   string  bar-value                  
  foo   string  /path/to/foo               
`

	err := p.printOutputsTable(outputs, "test-bundle")
	require.NoError(t, err)

	got := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, want, got)
}
