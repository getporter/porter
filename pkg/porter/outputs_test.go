package porter

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_printOutputsTable(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("testdata/show/object-parameter-value.json")
	require.NoError(t, err)
	var objVal map[string]interface{}
	err = json.Unmarshal(b, &objVal)
	require.NoError(t, err)

	p := NewTestPorter(t)
	defer p.Close()

	want := `---------------------------------------------------------------------------------
  Name     Type    Value                                                         
---------------------------------------------------------------------------------
  bar      string  ******                                                        
  foo      string  /path/to/foo                                                  
  object   object  {"a":{"b":1,"c":2},"d":"yay"}                                 
  longfoo  string  DFo6Wc2jDhmA7Yt4PbHyh8RO4vVG7leOzK412gf2TXNPJhuCUs1rB29nk...  
`

	outputs := DisplayValues{
		{Name: "bar", Type: "string", Value: "bar-value", Sensitive: true},
		{Name: "foo", Type: "string", Value: "/path/to/foo"},
		{Name: "object", Type: "object", Value: objVal},
		{Name: "longfoo", Type: "string", Value: "DFo6Wc2jDhmA7Yt4PbHyh8RO4vVG7leOzK412gf2TXNPJhuCUs1rB29nkJJd4ICimZGpyWpMGalSvDxf"},
	}
	err = p.printDisplayValuesTable(outputs)
	require.NoError(t, err)

	got := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, want, got)
}

func TestPorter_PrintBundleOutputs(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name           string
		format         printer.Format
		expectedOutput string
	}{
		{name: "text", format: printer.FormatPlaintext, expectedOutput: "testdata/outputs/show-expected-output.txt"},
		{name: "json", format: printer.FormatJson, expectedOutput: "testdata/outputs/show-expected-output.json"},
		{name: "yaml", format: printer.FormatYaml, expectedOutput: "testdata/outputs/show-expected-output.yaml"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

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
					"longfoo": &definition.Schema{
						Type: "string",
					},
					"porter-state": &definition.Schema{
						Type:    "string",
						Comment: "porter-internal", // This output should be hidden because it's internal
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
					"longfoo": {
						Definition: "longfoo",
					},
					"porter-state": {
						Definition: "porter-state",
						Path:       "/cnab/app/outputs/porter-state.tgz",
					},
				},
			}

			extB := cnab.NewBundle(b)
			i := p.TestInstallations.CreateInstallation(storage.NewInstallation("", "test"), func(i *storage.Installation) {
				i.Parameters.Parameters = p.SanitizeParameters(i.Parameters.Parameters, i.ID, extB)
			})
			c := p.TestInstallations.CreateRun(i.NewRun(cnab.ActionInstall, extB), func(r *storage.Run) {
				r.Bundle = b
				r.ParameterOverrides.Parameters = p.SanitizeParameters(r.ParameterOverrides.Parameters, r.ID, extB)
			})
			r := p.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))
			p.CreateOutput(r.NewOutput("foo", []byte("foo-output")), extB)
			p.CreateOutput(r.NewOutput("bar", []byte("bar-output")), extB)
			p.CreateOutput(r.NewOutput("longfoo", []byte("DFo6Wc2jDhmA7Yt4PbHyh8RO4vVG7leOzK412gf2TXNPJhuCUs1rB29nkJJd4ICimZGpyWpMGalSvDxf")), extB)
			p.CreateOutput(r.NewOutput("porter-state", []byte("porter-state.tgz contents")), extB)

			opts := OutputListOptions{
				installationOptions: installationOptions{
					Name: "test",
				},
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
			}
			err := p.PrintBundleOutputs(context.Background(), opts)
			require.NoError(t, err, "could not print bundle outputs")

			p.CompareGoldenFile(tc.expectedOutput, p.TestConfig.TestContext.GetOutput())
		})
	}
}

func TestOutputOptions_WithRunID(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name                string
		opts                interface{}
		runID              string
		installation       string
		expectedValidation bool
		expectedError      string
	}{
		{
			name: "OutputListOptions with RunID only",
			opts: &OutputListOptions{
				installationOptions: installationOptions{},
				PrintOptions: printer.PrintOptions{Format: printer.FormatPlaintext},
			},
			runID:              "01TEST_RUN_ID",
			expectedValidation: true,
		},
		{
			name: "OutputShowOptions with RunID only",
			opts: &OutputShowOptions{
				installationOptions: installationOptions{},
				Output: "test-output",
			},
			runID:              "01TEST_RUN_ID",
			expectedValidation: true,
		},
		{
			name: "OutputListOptions with both RunID and installation",
			opts: &OutputListOptions{
				installationOptions: installationOptions{
					Name: "test-installation",
				},
				PrintOptions: printer.PrintOptions{Format: printer.FormatPlaintext},
			},
			runID:              "01TEST_RUN_ID",
			expectedValidation: false,
			expectedError:      "either --installation or --run should be specified, not both",
		},
		{
			name: "OutputShowOptions with both RunID and installation", 
			opts: &OutputShowOptions{
				installationOptions: installationOptions{
					Name: "test-installation",
				},
				Output: "test-output",
			},
			runID:              "01TEST_RUN_ID",
			expectedValidation: false,
			expectedError:      "either --installation or --run should be specified, not both",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			switch opts := tc.opts.(type) {
			case *OutputListOptions:
				opts.RunID = tc.runID
				err := opts.Validate([]string{}, p.Context)
				if tc.expectedValidation {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tc.expectedError)
				}
			case *OutputShowOptions:
				opts.RunID = tc.runID
				err := opts.Validate([]string{"test-output"}, p.Context)
				if tc.expectedValidation {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tc.expectedError)
				}
			}
		})
	}
}

func TestPorter_ListBundleOutputs_WithRunID(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

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

	extB := cnab.NewBundle(b)
	i := p.TestInstallations.CreateInstallation(storage.NewInstallation("", "test"), func(i *storage.Installation) {
		i.Parameters.Parameters = p.SanitizeParameters(i.Parameters.Parameters, i.ID, extB)
	})
	c := p.TestInstallations.CreateRun(i.NewRun(cnab.ActionInstall, extB), func(r *storage.Run) {
		r.Bundle = b
		r.ParameterOverrides.Parameters = p.SanitizeParameters(r.ParameterOverrides.Parameters, r.ID, extB)
	})
	r := p.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))
	p.CreateOutput(r.NewOutput("foo", []byte("foo-output")), extB)
	p.CreateOutput(r.NewOutput("bar", []byte("bar-output")), extB)

	t.Run("ListBundleOutputs with RunID", func(t *testing.T) {
		opts := &OutputListOptions{
			installationOptions: installationOptions{},
			PrintOptions: printer.PrintOptions{
				Format: printer.FormatPlaintext,
			},
			RunID: c.ID,
		}

		outputs, err := p.ListBundleOutputs(context.Background(), opts)
		require.NoError(t, err, "ListBundleOutputs with RunID should not fail")
		assert.Equal(t, 2, len(outputs), "should return correct number of outputs")

		// Find outputs by name since order might vary
		var barOutput, fooOutput *DisplayValue
		for i := range outputs {
			if outputs[i].Name == "bar" {
				barOutput = &outputs[i]
			} else if outputs[i].Name == "foo" {
				fooOutput = &outputs[i]
			}
		}

		require.NotNil(t, barOutput, "should have found bar output")
		assert.Equal(t, "bar", barOutput.Name)
		assert.Equal(t, "bar-output", barOutput.Value)

		require.NotNil(t, fooOutput, "should have found foo output")
		assert.Equal(t, "foo", fooOutput.Name) 
		assert.True(t, fooOutput.Sensitive, "foo output should be marked as sensitive")
		assert.Equal(t, "foo-output", fooOutput.Value) // The actual value should be stored
		assert.Equal(t, "******", fooOutput.PrintValue()) // But should be masked when printed
	})
}

func TestPorter_ShowBundleOutput_WithRunID(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	// Create test data
	b := bundle.Bundle{
		Definitions: definition.Definitions{
			"test-output": &definition.Schema{
				Type: "string",
			},
		},
		Outputs: map[string]bundle.Output{
			"test-output": {
				Definition: "test-output",
			},
		},
	}

	extB := cnab.NewBundle(b)
	i := p.TestInstallations.CreateInstallation(storage.NewInstallation("", "test"), func(i *storage.Installation) {
		i.Parameters.Parameters = p.SanitizeParameters(i.Parameters.Parameters, i.ID, extB)
	})
	c := p.TestInstallations.CreateRun(i.NewRun(cnab.ActionInstall, extB), func(r *storage.Run) {
		r.Bundle = b
		r.ParameterOverrides.Parameters = p.SanitizeParameters(r.ParameterOverrides.Parameters, r.ID, extB)
	})
	r := p.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))
	p.CreateOutput(r.NewOutput("test-output", []byte("test-value")), extB)

	t.Run("ShowBundleOutput with RunID", func(t *testing.T) {
		opts := &OutputShowOptions{
			installationOptions: installationOptions{},
			Output:              "test-output",
			RunID:               c.ID,
		}

		err := p.ShowBundleOutput(context.Background(), opts)
		require.NoError(t, err, "ShowBundleOutput with RunID should not fail")

		output := p.TestConfig.TestContext.GetOutput()
		assert.Equal(t, "test-value", output)
	})
}
