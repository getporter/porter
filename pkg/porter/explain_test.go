package porter

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExplain_generateActionsTableNoActions(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)

	p.printActionsExplainTable(&bun)
	expected := "Name   Description   Modifies Installation   Stateless\n"

	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateActionsBlockNoActions(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)

	p.printActionsExplainBlock(&bun)
	expected := "No custom actions defined\n\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateCredentialsTableNoCreds(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)

	p.printCredentialsExplainTable(&bun)
	expected := "Name   Description   Required\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateCredentialsBlockNoCreds(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)

	p.printCredentialsExplainBlock(&bun)
	expected := "No credentials defined\n\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateOutputsTableNoOutputs(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)

	p.printOutputsExplainTable(&bun)
	expected := "Name   Description   Type   Applies To\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateOutputsBlockNoOutputs(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)

	p.printOutputsExplainBlock(&bun)
	expected := "No outputs defined\n\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateParametersTableNoParams(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)

	p.printParametersExplainTable(&bun)
	expected := "Name   Description   Type   Default   Required   Applies To\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateParametersBlockNoParams(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)

	p.printParametersExplainBlock(&bun)
	expected := "No parameters defined\n\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_validateBadFormat(t *testing.T) {

	p := NewTestPorter(t)

	opts := ExplainOpts{}
	opts.RawFormat = "vpml"

	err := opts.Validate([]string{}, p.Context)
	assert.EqualError(t, err, "invalid format: vpml")
}

func TestExplain_generateTable(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.TestContext.AddTestFile("testdata/explain/params-bundle.json", "params-bundle.json")
	b, err := p.CNAB.LoadBundle("params-bundle.json")

	pb, err := generatePrintable(b)
	require.NoError(t, err)
	opts := ExplainOpts{}
	opts.RawFormat = "table"

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.printBundleExplain(opts, pb)
	assert.NoError(t, err)
	gotOutput := p.TestConfig.TestContext.GetOutput()
	expected, err := ioutil.ReadFile("testdata/explain/expected-table-output.txt")
	require.NoError(t, err)
	assert.Equal(t, string(expected), gotOutput)
}

func TestExplain_generateJSON(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.TestContext.AddTestFile("testdata/explain/params-bundle.json", "params-bundle.json")
	b, err := p.CNAB.LoadBundle("params-bundle.json")

	pb, err := generatePrintable(b)
	require.NoError(t, err)
	opts := ExplainOpts{}
	opts.RawFormat = "json"

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.printBundleExplain(opts, pb)
	assert.NoError(t, err)
	gotOutput := p.TestConfig.TestContext.GetOutput()
	expected, err := ioutil.ReadFile("testdata/explain/expected-json-output.json")
	require.NoError(t, err)
	assert.Equal(t, string(expected), gotOutput)
}

func TestExplain_generateYAML(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.TestContext.AddTestFile("testdata/explain/params-bundle.json", "params-bundle.json")
	b, err := p.CNAB.LoadBundle("params-bundle.json")

	pb, err := generatePrintable(b)
	require.NoError(t, err)

	opts := ExplainOpts{}
	opts.RawFormat = "yaml"

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.printBundleExplain(opts, pb)
	assert.NoError(t, err)
	gotOutput := p.TestConfig.TestContext.GetOutput()
	expected, err := ioutil.ReadFile("testdata/explain/expected-yaml-output.yaml")
	require.NoError(t, err)
	assert.Equal(t, string(expected), gotOutput)
}

func TestExplain_generatePrintableBundleParams(t *testing.T) {
	bun := &bundle.Bundle{
		Definitions: definition.Definitions{
			"string": &definition.Schema{
				Type:    "string",
				Default: "clippy",
			},
		},
		Parameters: map[string]bundle.Parameter{
			"debug": {
				Definition: "string",
				Required:   true,
			},
		},
	}

	pb, err := generatePrintable(bun)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(pb.Parameters))
	d := pb.Parameters[0]
	assert.Equal(t, "clippy", fmt.Sprintf("%v", d.Default))
	assert.Equal(t, 0, len(pb.Outputs))
	assert.Equal(t, 0, len(pb.Credentials))
	assert.Equal(t, 0, len(pb.Actions))
}

func TestExplain_generatePrintableBundleOutputs(t *testing.T) {
	bun := &bundle.Bundle{
		Definitions: definition.Definitions{
			"string": &definition.Schema{
				Type:    "string",
				Default: "clippy",
			},
		},
		Outputs: map[string]bundle.Output{
			"debug": {
				Definition: "string",
			},
		},
	}

	pb, err := generatePrintable(bun)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(pb.Outputs))
	d := pb.Outputs[0]
	assert.Equal(t, "string", fmt.Sprintf("%v", d.Type))
	assert.Equal(t, 0, len(pb.Parameters))
	assert.Equal(t, 0, len(pb.Credentials))
	assert.Equal(t, 0, len(pb.Actions))
}

func TestExplain_generatePrintableBundleCreds(t *testing.T) {
	bun := &bundle.Bundle{
		Credentials: map[string]bundle.Credential{
			"debug": {
				Required:    true,
				Description: "a cred",
			},
		},
	}

	pb, err := generatePrintable(bun)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(pb.Credentials))
	d := pb.Credentials[0]
	assert.True(t, d.Required)
	assert.Equal(t, "a cred", d.Description)
	assert.Equal(t, 0, len(pb.Parameters))
	assert.Equal(t, 0, len(pb.Outputs))
	assert.Equal(t, 0, len(pb.Actions))
}

func TestExplain_genratePrintablBundle_empty(t *testing.T) {
	var bun *bundle.Bundle
	_, err := generatePrintable(bun)
	assert.Error(t, err)
	assert.EqualError(t, err, "expected a bundle")
}
