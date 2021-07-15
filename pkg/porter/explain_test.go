package porter

import (
	"fmt"
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/cnab/extensions"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExplain_generateActionsTableNoActions(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)
	defer p.Teardown()

	p.printActionsExplainTable(&bun)
	expected := "Name   Description   Modifies Installation   Stateless\n"

	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateActionsBlockNoActions(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)
	defer p.Teardown()

	p.printActionsExplainBlock(&bun)
	expected := "No custom actions defined\n\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateCredentialsTableNoCreds(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)
	defer p.Teardown()

	p.printCredentialsExplainTable(&bun)
	expected := "Name   Description   Required   Applies To\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateCredentialsBlockNoCreds(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)
	defer p.Teardown()

	p.printCredentialsExplainBlock(&bun)
	expected := "No credentials defined\n\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateOutputsTableNoOutputs(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)
	defer p.Teardown()

	p.printOutputsExplainTable(&bun)
	expected := "Name   Description   Type   Applies To\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateOutputsBlockNoOutputs(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)
	defer p.Teardown()

	p.printOutputsExplainBlock(&bun)
	expected := "No outputs defined\n\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateParametersTableNoParams(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)
	defer p.Teardown()

	p.printParametersExplainTable(&bun)
	expected := "Name   Description   Type   Default   Required   Applies To\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_generateParametersBlockNoParams(t *testing.T) {
	bun := PrintableBundle{}

	p := NewTestPorter(t)
	defer p.Teardown()

	p.printParametersExplainBlock(&bun)
	expected := "No parameters defined\n\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, expected, gotOutput)
	t.Log(gotOutput)
}

func TestExplain_validateBadFormat(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	opts := ExplainOpts{}
	opts.RawFormat = "vpml"

	err := opts.Validate([]string{}, p.Context)
	assert.EqualError(t, err, "invalid format: vpml")
}

func TestExplain_generateTable(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/explain/params-bundle.json", "params-bundle.json")
	b, err := p.CNAB.LoadBundle("params-bundle.json")

	pb, err := generatePrintable(b, "")
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
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/explain/params-bundle.json", "params-bundle.json")
	b, err := p.CNAB.LoadBundle("params-bundle.json")

	pb, err := generatePrintable(b, "")
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
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/explain/params-bundle.json", "params-bundle.json")
	b, err := p.CNAB.LoadBundle("params-bundle.json")

	pb, err := generatePrintable(b, "")
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
	bun := bundle.Bundle{
		RequiredExtensions: []string{
			extensions.FileParameterExtensionKey,
		},
		Definitions: definition.Definitions{
			"string": &definition.Schema{
				Type:    "string",
				Default: "clippy",
			},
			"file": &definition.Schema{
				Type:            "string",
				ContentEncoding: "base64",
			},
		},
		Parameters: map[string]bundle.Parameter{
			"debug": {
				Definition: "string",
				Required:   true,
			},
			"tfstate": {
				Definition: "file",
			},
		},
		Custom: map[string]interface{}{
			"sh.porter": map[string]interface{}{
				"manifest": "5040d45d0c44e7632563966c33f5e8980e83cfa7c0485f725b623b7604f072f0",
				"version":  "v0.30.0",
				"commit":   "3b7c85ba",
			},
		},
	}

	pb, err := generatePrintable(bun, "")
	require.NoError(t, err)

	require.Equal(t, 2, len(pb.Parameters), "expected 2 parameters")
	d := pb.Parameters[0]
	require.Equal(t, "debug", d.Name)
	assert.Equal(t, "clippy", fmt.Sprintf("%v", d.Default))
	assert.Equal(t, "string", d.Type)
	f := pb.Parameters[1]
	require.Equal(t, "tfstate", f.Name)
	assert.Equal(t, "file", f.Type)

	assert.Equal(t, 0, len(pb.Outputs))
	assert.Equal(t, 0, len(pb.Credentials))
	assert.Equal(t, 0, len(pb.Actions))
}

func TestExplain_generatePrintableBundleParamsWithAction(t *testing.T) {
	bun := bundle.Bundle{
		RequiredExtensions: []string{
			extensions.FileParameterExtensionKey,
		},
		Definitions: definition.Definitions{
			"string": &definition.Schema{
				Type:    "string",
				Default: "clippy",
			},
			"file": &definition.Schema{
				Type:            "string",
				ContentEncoding: "base64",
			},
		},
		Parameters: map[string]bundle.Parameter{
			"debug": {
				Definition: "string",
				Required:   true,
				ApplyTo:    []string{"install"},
			},
			"tfstate": {
				Definition: "file",
			},
		},
		Custom: map[string]interface{}{
			"sh.porter": map[string]interface{}{
				"manifest": "5040d45d0c44e7632563966c33f5e8980e83cfa7c0485f725b623b7604f072f0",
				"version":  "v0.30.0",
				"commit":   "3b7c85ba",
			},
		},
	}

	t.Run("action applies", func(t *testing.T) {
		pb, err := generatePrintable(bun, "install")
		require.NoError(t, err)

		require.Equal(t, 2, len(pb.Parameters), "expected 2 parameters")

		d := pb.Parameters[0]
		require.Equal(t, "debug", d.Name)
		assert.Equal(t, "install", d.ApplyTo)

		f := pb.Parameters[1]
		require.Equal(t, "tfstate", f.Name)
		assert.Equal(t, "All Actions", f.ApplyTo)
	})

	t.Run("action does not apply", func(t *testing.T) {
		pb, err := generatePrintable(bun, "upgrade")
		require.NoError(t, err)

		require.Equal(t, 1, len(pb.Parameters), "expected only 1 parameter since debug parameter doesn't apply to upgrade command")
		require.Equal(t, "tfstate", pb.Parameters[0].Name)
	})

	t.Run("all actions", func(t *testing.T) {
		pb, err := generatePrintable(bun, "")
		require.NoError(t, err)

		require.Equal(t, 2, len(pb.Parameters), "expected 2 parameters")

		d := pb.Parameters[0]
		require.Equal(t, "debug", d.Name)
		assert.Equal(t, "install", d.ApplyTo)

		f := pb.Parameters[1]
		require.Equal(t, "tfstate", f.Name)
		assert.Equal(t, "All Actions", f.ApplyTo)
	})
}

func TestExplain_generatePrintableBundleOutputs(t *testing.T) {
	bun := bundle.Bundle{
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
			"someoutput": {
				Definition: "string",
				ApplyTo:    []string{"install"},
			},
		},
		Custom: map[string]interface{}{
			"sh.porter": map[string]interface{}{
				"manifest": "5040d45d0c44e7632563966c33f5e8980e83cfa7c0485f725b623b7604f072f0",
				"version":  "v0.30.0",
				"commit":   "3b7c85ba",
			},
		},
	}

	pb, err := generatePrintable(bun, "")
	require.NoError(t, err)

	require.Equal(t, 2, len(pb.Outputs), "expected someoutput to be included because the action is unset")
	debugOutput := pb.Outputs[0]
	assert.Equal(t, "string", fmt.Sprintf("%v", debugOutput.Type))
	assert.Equal(t, "debug", debugOutput.Name)
	assert.Equal(t, "All Actions", debugOutput.ApplyTo)

	someOutput := pb.Outputs[1]
	assert.Equal(t, "string", fmt.Sprintf("%v", someOutput.Type))
	assert.Equal(t, "someoutput", someOutput.Name)
	assert.Equal(t, "install", someOutput.ApplyTo)

	assert.Equal(t, 0, len(pb.Parameters))
	assert.Equal(t, 0, len(pb.Credentials))
	assert.Equal(t, 0, len(pb.Actions))

	// Check outputs for install action
	pb, err = generatePrintable(bun, "install")
	require.NoError(t, err)
	assert.Equal(t, 2, len(pb.Outputs), "expected someoutput to be included")

	// Check outputs for upgrade action action (someoutput doesn't apply)
	pb, err = generatePrintable(bun, "upgrade")
	require.NoError(t, err)
	assert.Equal(t, 1, len(pb.Outputs), "expected someoutput to be excluded by its applyTo")
}

func TestExplain_generatePrintableBundleCreds(t *testing.T) {
	bun := bundle.Bundle{
		Credentials: map[string]bundle.Credential{
			"kubeconfig": {
				Required:    true,
				Description: "a cred",
				ApplyTo:     []string{"install"},
			},
			"password": {
				Description: "another cred",
			},
		},
		Custom: map[string]interface{}{
			"sh.porter": map[string]interface{}{
				"manifest": "5040d45d0c44e7632563966c33f5e8980e83cfa7c0485f725b623b7604f072f0",
				"version":  "v0.30.0",
				"commit":   "3b7c85ba",
			},
		},
	}

	t.Run("action applies", func(t *testing.T) {
		pb, err := generatePrintable(bun, "install")
		require.NoError(t, err)

		require.Equal(t, 2, len(pb.Credentials), "expected 2 credentials")

		d := pb.Credentials[0]
		require.Equal(t, "kubeconfig", d.Name)
		assert.Equal(t, "install", d.ApplyTo)

		f := pb.Credentials[1]
		require.Equal(t, "password", f.Name)
		assert.Equal(t, "All Actions", f.ApplyTo)
	})

	t.Run("action does not apply", func(t *testing.T) {
		pb, err := generatePrintable(bun, "upgrade")
		require.NoError(t, err)

		require.Equal(t, 1, len(pb.Credentials), "expected only 1 credential since kubeconfig credential doesn't apply to upgrade command")
		require.Equal(t, "password", pb.Credentials[0].Name)
	})

	t.Run("all actions", func(t *testing.T) {
		pb, err := generatePrintable(bun, "")
		require.NoError(t, err)

		require.Equal(t, 2, len(pb.Credentials), "expected 2 credentials")

		d := pb.Credentials[0]
		require.Equal(t, "kubeconfig", d.Name)
		assert.Equal(t, "install", d.ApplyTo)

		f := pb.Credentials[1]
		require.Equal(t, "password", f.Name)
		assert.Equal(t, "All Actions", f.ApplyTo)
	})
}

func TestExplain_generatePrintableBundlePorterVersion(t *testing.T) {
	bun := bundle.Bundle{
		Definitions: definition.Definitions{
			"string": &definition.Schema{
				Type:    "string",
				Default: "clippy",
			},
		},
		Custom: map[string]interface{}{
			"sh.porter": map[string]interface{}{
				"manifest": "5040d45d0c44e7632563966c33f5e8980e83cfa7c0485f725b623b7604f072f0",
				"version":  "v0.30.0",
				"commit":   "3b7c85ba",
			},
		},
	}

	pb, err := generatePrintable(bun, "")
	assert.NoError(t, err)

	assert.Equal(t, "v0.30.0", pb.PorterVersion)
}

func TestExplain_generatePrintableBundlePorterVersionNonPorterBundle(t *testing.T) {
	bun := bundle.Bundle{
		Definitions: definition.Definitions{
			"string": &definition.Schema{
				Type:    "string",
				Default: "clippy",
			},
		},
	}

	pb, err := generatePrintable(bun, "")
	assert.NoError(t, err)

	assert.Equal(t, "", pb.PorterVersion)
}

func TestExplain_generatePrintableBundleDependencies(t *testing.T) {

	sequenceMock := []string{"nginx", "storage", "mysql"}
	bun := bundle.Bundle{
		Custom: map[string]interface{}{
			extensions.DependenciesExtensionKey: extensions.Dependencies{
				Sequence: sequenceMock,
				Requires: map[string]extensions.Dependency{
					"mysql": extensions.Dependency{
						Name:   "mysql",
						Bundle: "somecloud/mysql:0.1.0",
					},
					"storage": extensions.Dependency{
						Name:   "storage",
						Bundle: "localhost:5000/blob-storage:0.1.0",
					},
					"nginx": extensions.Dependency{
						Name:   "nginx",
						Bundle: "localhost:5000/nginx:1.19",
					},
				},
			},
			"sh.porter": map[string]interface{}{
				"manifest": "5040d45d0c44e7632563966c33f5e8980e83cfa7c0485f725b623b7604f072f0",
				"version":  "v0.30.0",
				"commit":   "3b7c85ba",
			},
		},
	}

	pd, err := generatePrintable(bun, "")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(pd.Dependencies))
	assert.Equal(t, 0, len(pd.Parameters))
	assert.Equal(t, 0, len(pd.Outputs))
	assert.Equal(t, 0, len(pd.Actions))
	assert.Equal(t, "nginx", pd.Dependencies[0].Alias)
	assert.Equal(t, "somecloud/mysql:0.1.0", pd.Dependencies[2].Reference)
}

func TestExplain_generateJSONForDependencies(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/explain/dependencies-bundle.json", "dependencies-bundle.json")
	b, err := p.CNAB.LoadBundle("dependencies-bundle.json")

	pb, err := generatePrintable(b, "")
	require.NoError(t, err)
	opts := ExplainOpts{}
	opts.RawFormat = "json"

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.printBundleExplain(opts, pb)
	assert.NoError(t, err)
	gotOutput := p.TestConfig.TestContext.GetOutput()
	expected, err := ioutil.ReadFile("testdata/explain/expected-json-dependencies-output.json")
	require.NoError(t, err)
	assert.Equal(t, string(expected), gotOutput)
}
