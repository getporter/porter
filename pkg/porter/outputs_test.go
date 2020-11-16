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

func TestPorter_printOutputsTable(t *testing.T) {
	p := NewTestPorter(t)

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

	c := p.TestClaims.CreateClaim("test", claim.ActionInstall, b, nil)
	r := p.TestClaims.CreateResult(c, claim.StatusSucceeded)
	p.TestClaims.CreateOutput(c, r, "foo", []byte("foo-output"))
	p.TestClaims.CreateOutput(c, r, "bar", []byte("bar-output"))

	// Hard code the date so we can compare the command output easily
	c.Created = time.Date(1983, time.April, 18, 1, 2, 3, 4, time.UTC)
	err := p.TestClaims.SaveClaim(c)
	require.NoError(t, err, "could not store claim")

	opts := OutputListOptions{
		sharedOptions: sharedOptions{
			Name: "test",
		},
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatJson,
		},
	}
	err = p.PrintBundleOutputs(opts)
	require.NoError(t, err, "could not print bundle outputs")

	want := `[
  {
    "Name": "bar",
    "Value": "bar-output",
    "Type": "string"
  },
  {
    "Name": "foo",
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

	fullOutputValue := "this-lengthy-output-will-be-truncated-if-the-output-format-is-table"

	b := bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type: "string",
			},
		},
		Outputs: map[string]bundle.Output{
			"foo": {
				Definition: "foo",
			},
		},
	}

	c, err := claim.New("test", claim.ActionInstall, b, nil)
	c.Action = claim.ActionInstall
	require.NoError(t, err, "NewClaim failed")

	err = p.Claims.SaveClaim(c)
	require.NoError(t, err, "SaveClaim failed")

	r, err := c.NewResult(claim.StatusSucceeded)
	require.NoError(t, err, "NewResult failed")
	err = p.Claims.SaveResult(r)
	require.NoError(t, err, "SaveResult failed")

	foo := claim.NewOutput(c, r, "foo", []byte(fullOutputValue))
	err = p.Claims.SaveOutput(foo)
	require.NoError(t, err, "SaveOutput failed")

	testcases := []struct {
		name          string
		opts          OutputListOptions
		expectedValue string
	}{
		{
			"format Table",
			OutputListOptions{
				sharedOptions: sharedOptions{Name: "test"},
				PrintOptions:  printer.PrintOptions{Format: printer.FormatTable},
			},
			"this-lengthy-output-will-be-truncated-if-the-output-forma...",
		},
		{
			"format YAML",
			OutputListOptions{
				sharedOptions: sharedOptions{Name: "test"},
				PrintOptions:  printer.PrintOptions{Format: printer.FormatYaml},
			},
			fullOutputValue,
		},
		{
			"format JSON",
			OutputListOptions{
				sharedOptions: sharedOptions{Name: "test"},
				PrintOptions:  printer.PrintOptions{Format: printer.FormatJson},
			},
			fullOutputValue,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gotOutputs, err := p.ListBundleOutputs(&tc.opts)
			require.NoError(t, err, "ListBundleOutputs failed")

			wantOutputs := DisplayOutputs{
				{
					Name:  "foo",
					Type:  "string",
					Value: tc.expectedValue,
				},
			}
			require.Equal(t, wantOutputs, gotOutputs)
		})
	}
}
