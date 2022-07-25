package porter

import (
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	depsv1 "get.porter.sh/porter/pkg/cnab/dependencies/v1"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/test"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExplain_ValidateReference(t *testing.T) {
	const ref = "ghcr.io/getporter/examples/porter-hello:v0.2.0"
	t.Run("--reference specified", func(t *testing.T) {
		porterCtx := portercontext.NewTestContext(t)
		opts := ExplainOpts{}
		opts.Reference = ref

		err := opts.Validate(nil, porterCtx.Context)
		require.NoError(t, err, "Validate failed")
		assert.Equal(t, ref, opts.Reference)
	})
	t.Run("reference positional argument specified", func(t *testing.T) {
		porterCtx := portercontext.NewTestContext(t)
		opts := ExplainOpts{}

		err := opts.Validate([]string{ref}, porterCtx.Context)
		require.NoError(t, err, "Validate failed")
		assert.Equal(t, ref, opts.Reference)
	})
}

func TestExplain_validateBadFormat(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := ExplainOpts{}
	opts.RawFormat = "vpml"

	err := opts.Validate([]string{}, p.Context)
	assert.EqualError(t, err, "invalid format: vpml")
}

func TestExplain_generateTable(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/explain/params-bundle.json", "params-bundle.json")
	b, err := p.CNAB.LoadBundle("params-bundle.json")
	require.NoError(t, err)

	pb, err := generatePrintable(b, "")
	require.NoError(t, err)
	opts := ExplainOpts{}
	opts.RawFormat = "plaintext"

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.printBundleExplain(opts, pb, b)
	assert.NoError(t, err)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	test.CompareGoldenFile(t, "testdata/explain/expected-table-output.txt", gotOutput)
}

func TestExplain_generateTableRequireDockerHostAccess(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/explain/bundle-docker.json", "bundle-docker.json")
	b, err := p.CNAB.LoadBundle("bundle-docker.json")
	require.NoError(t, err)

	pb, err := generatePrintable(b, "")
	require.NoError(t, err)
	opts := ExplainOpts{}
	opts.RawFormat = "plaintext"

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.printBundleExplain(opts, pb, b)
	assert.NoError(t, err)
	gotOutput := p.TestConfig.TestContext.GetOutput()
	p.CompareGoldenFile("testdata/explain/expected-table-output-docker.txt", gotOutput)
}

func TestExplain_generateJSON(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/explain/params-bundle.json", "params-bundle.json")
	b, err := p.CNAB.LoadBundle("params-bundle.json")
	require.NoError(t, err)

	pb, err := generatePrintable(b, "")
	require.NoError(t, err)
	opts := ExplainOpts{}
	opts.RawFormat = "json"

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.printBundleExplain(opts, pb, b)
	assert.NoError(t, err)
	gotOutput := p.TestConfig.TestContext.GetOutput()
	p.CompareGoldenFile("testdata/explain/expected-json-output.json", gotOutput)
}

func TestExplain_generateYAML(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/explain/params-bundle.json", "params-bundle.json")
	b, err := p.CNAB.LoadBundle("params-bundle.json")
	require.NoError(t, err)

	pb, err := generatePrintable(b, "")
	require.NoError(t, err)

	opts := ExplainOpts{}
	opts.RawFormat = "yaml"

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.printBundleExplain(opts, pb, b)
	assert.NoError(t, err)
	gotOutput := p.TestConfig.TestContext.GetOutput()
	p.CompareGoldenFile("testdata/explain/expected-yaml-output.yaml", gotOutput)
}

func TestExplain_generatePrintableBundleParams(t *testing.T) {
	bun := cnab.NewBundle(bundle.Bundle{
		RequiredExtensions: []string{
			cnab.FileParameterExtensionKey,
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
	})

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
	bun := cnab.NewBundle(bundle.Bundle{
		RequiredExtensions: []string{
			cnab.FileParameterExtensionKey,
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
	})

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
	bun := cnab.NewBundle(bundle.Bundle{
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
	})

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
	bun := cnab.NewBundle(bundle.Bundle{
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
	})

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
	bun := cnab.NewBundle(bundle.Bundle{
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
	})

	pb, err := generatePrintable(bun, "")
	assert.NoError(t, err)

	assert.Equal(t, "v0.30.0", pb.PorterVersion)
}

func TestExplain_generatePrintableBundlePorterVersionNonPorterBundle(t *testing.T) {
	bun := cnab.NewBundle(bundle.Bundle{
		Definitions: definition.Definitions{
			"string": &definition.Schema{
				Type:    "string",
				Default: "clippy",
			},
		},
	})

	pb, err := generatePrintable(bun, "")
	assert.NoError(t, err)

	assert.Equal(t, "", pb.PorterVersion)
}

func TestExplain_generatePrintableBundleDependencies(t *testing.T) {
	sequenceMock := []string{"nginx", "storage", "mysql"}
	bun := cnab.NewBundle(bundle.Bundle{
		Custom: map[string]interface{}{
			cnab.DependenciesV1ExtensionKey: depsv1.Dependencies{
				Sequence: sequenceMock,
				Requires: map[string]depsv1.Dependency{
					"mysql": {
						Name:   "mysql",
						Bundle: "somecloud/mysql:0.1.0",
					},
					"storage": {
						Name:   "storage",
						Bundle: "localhost:5000/blob-storage:0.1.0",
					},
					"nginx": {
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
	})

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
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/explain/dependencies-bundle.json", "dependencies-bundle.json")
	b, err := p.CNAB.LoadBundle("dependencies-bundle.json")
	require.NoError(t, err)

	pb, err := generatePrintable(b, "")
	require.NoError(t, err)
	opts := ExplainOpts{}
	opts.RawFormat = "json"

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.printBundleExplain(opts, pb, b)
	assert.NoError(t, err)
	gotOutput := p.TestConfig.TestContext.GetOutput()

	p.CompareGoldenFile("testdata/explain/expected-json-dependencies-output.json", gotOutput)
}

func TestExplain_generateTableNonPorterBundle(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/explain/params-bundle-non-porter.json", "params-bundle.json")
	b, err := p.CNAB.LoadBundle("params-bundle.json")
	require.NoError(t, err)

	pb, err := generatePrintable(b, "")
	require.NoError(t, err)
	opts := ExplainOpts{}
	opts.RawFormat = "plaintext"

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.printBundleExplain(opts, pb, b)
	assert.NoError(t, err)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	test.CompareGoldenFile(t, "testdata/explain/expected-table-output-non-porter.txt", gotOutput)
}

func TestExplain_generateTableBundleWithNoMixins(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/explain/params-bundle-no-mixins.json", "params-bundle.json")
	b, err := p.CNAB.LoadBundle("params-bundle.json")
	require.NoError(t, err)

	pb, err := generatePrintable(b, "")
	require.NoError(t, err)
	opts := ExplainOpts{}
	opts.RawFormat = "plaintext"

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.printBundleExplain(opts, pb, b)
	assert.NoError(t, err)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	test.CompareGoldenFile(t, "testdata/explain/expected-table-output-no-mixins.txt", gotOutput)
}
