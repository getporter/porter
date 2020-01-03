package porter

import (
	"testing"
	"time"

	"get.porter.sh/porter/pkg/printer"
	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

func TestPorter_printOutputsTable(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.CNAB = NewTestCNABProvider()

	want := `------------------------------
  Name  Type    Value         
------------------------------
  bar   string  bar-value     
  foo   string  /path/to/foo  
`

	outputs := []DisplayOutput{
		{Name: "bar", Type: "string", Value: "bar-value"},
		{Name: "foo", Type: "string", Value: "/path/to/foo"},
	}
	err := p.printOutputsTable(outputs)
	require.NoError(t, err)

	got := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, want, got)
}

func TestPorter_printDisplayOutput_JSON(t *testing.T) {
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

	err := p.InstanceStorage.Store(claim)
	require.NoError(t, err, "could not store claim")

	opts := OutputListOptions{
		sharedOptions: sharedOptions{
			Name: "test-bundle",
		},
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatJson,
		},
	}
	err = p.PrintBundleOutputs(&opts)
	require.NoError(t, err, "could not print bundle outputs")

	want := `[
  {
    "Name": "bar",
    "Definition": {
      "type": "string"
    },
    "Value": "bar-output",
    "Type": "string"
  },
  {
    "Name": "foo",
    "Definition": {
      "type": "string",
      "writeOnly": true
    },
    "Value": "foo-output",
    "Type": "string"
  }
]
`

	got := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, want, got)
}

func TestPorter_ListOutputs_Truncation(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.CNAB = NewTestCNABProvider()

	fullOutputValue := "this-lengthy-output-will-be-truncated-if-the-output-format-is-table"

	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Bundle = &bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type: "string",
			},
		},
		Outputs: map[string]bundle.Output{
			"foo": bundle.Output{
				Definition: "foo",
			},
		},
	}
	claim.Outputs = map[string]interface{}{
		"foo": fullOutputValue,
	}

	testcases := []struct {
		name          string
		format        printer.Format
		expectedValue string
	}{
		{
			"format Table",
			printer.FormatTable,
			"this-lengthy-output-will-be-truncated-if-the-output-forma...",
		},
		{
			"format YAML",
			printer.FormatYaml,
			fullOutputValue,
		},
		{
			"format JSON",
			printer.FormatJson,
			fullOutputValue,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gotOutputs := p.ListBundleOutputs(*claim, tc.format)

			wantOutputs := []DisplayOutput{
				{
					Name:       "foo",
					Definition: *claim.Bundle.Definitions["foo"],
					Type:       "string",
					Value:      tc.expectedValue,
				},
			}
			require.Equal(t, wantOutputs, gotOutputs)
		})
	}
}
