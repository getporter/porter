//go:build integration

package integration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	testhelper "get.porter.sh/porter/pkg/test"
	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func TestSchema(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Only use a well-known set of mixin in our home directory so that the schema is consistent
	mixinsDir := filepath.Join(test.PorterHomeDir, "mixins")
	mixins, err := os.ReadDir(mixinsDir)
	for _, mixinDir := range mixins {
		mixinName := mixinDir.Name()
		if mixinName != "exec" && mixinName != "testmixin" {
			require.NoError(t, os.RemoveAll(filepath.Join(mixinsDir, mixinName)))
		}
	}

	// Validate that the schema matches what we expected
	gotSchema, _ := test.RequirePorter("schema")
	testhelper.CompareGoldenFile(t, "testdata/schema/schema.json", gotSchema)

	testManifests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "valid", path: filepath.Join(test.RepoRoot, "tests/testdata/mybuns/porter.yaml")},
		{name: "invalid mixin declarations", path: "testdata/schema/invalid_mixin_config.yaml",
			wantErr: `mixins.1: mixins.1 must be one of the following: "exec", "testmixin"
mixins.2: Must validate one and only one schema (oneOf)
mixins.2.testmixin: Additional property missingproperty is not allowed`},
	}

	// Validate each manifest against the schema
	for _, tm := range testManifests {
		t.Run(tm.name, func(t *testing.T) {
			// Load the manifest as a go dump
			testManifestPath := tm.path
			testManifest, err := ioutil.ReadFile(testManifestPath)
			require.NoError(t, err, "failed to read %s", testManifestPath)

			m := make(map[string]interface{})
			err = yaml.Unmarshal(testManifest, &m)
			require.NoError(t, err, "failed to unmarshal %s", testManifestPath)

			// Load the manifest schema returned from `porter schema`
			manifestLoader := gojsonschema.NewGoLoader(m)
			schemaLoader := gojsonschema.NewStringLoader(gotSchema)

			// Validate the manifest against the schema
			fails, err := gojsonschema.Validate(schemaLoader, manifestLoader)
			require.NoError(t, err)

			if tm.wantErr == "" {
				assert.Empty(t, fails.Errors(), "expected %s to validate against the porter schema", testManifestPath)
				// Print any validation errors returned
				for _, err := range fails.Errors() {
					t.Logf("%s", err)
				}
			} else {
				var allFails strings.Builder
				for _, err := range fails.Errors() {
					allFails.WriteString(err.String())
					allFails.WriteString("\n")
				}
				tests.RequireOutputContains(t, allFails.String(), tm.wantErr)
			}
		})
	}
}
