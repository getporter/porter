package porter

import (
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml" // We are not using go-yaml because of serialization problems with jsonschema, don't use this library elsewhere
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func TestPorter_PrintManifestSchema(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	err := p.PrintManifestSchema()
	require.NoError(t, err)

	p.CompareGoldenFile("testdata/schema.json", p.TestConfig.TestContext.GetOutput())
}

func TestPorter_ValidateManifestSchema(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	// Load the default Porter manifest
	b, err := ioutil.ReadFile("testdata/porter.yaml")
	require.NoError(t, err)

	// Load the manifest as a go dump
	m := make(map[string]interface{})
	err = yaml.Unmarshal(b, &m)
	require.NoError(t, err)
	manifestLoader := gojsonschema.NewGoLoader(m)

	// Load the manifest schema
	err = p.PrintManifestSchema()
	require.NoError(t, err, "could not generate schema")
	schema := p.TestConfig.TestContext.GetOutput()
	schemaLoader := gojsonschema.NewStringLoader(schema)

	// Validate the manifest against the schema
	fails, err := gojsonschema.Validate(schemaLoader, manifestLoader)
	require.NoError(t, err)

	assert.Empty(t, fails.Errors(), "expected testdata/porter.yaml to validate against the porter schema")
	// Print it pretty like
	for _, err := range fails.Errors() {
		t.Logf("%s", err)
	}
}
