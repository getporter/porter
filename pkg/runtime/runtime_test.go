package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorterRuntime_Execute_readMixinOutputs(t *testing.T) {
	r := NewTestPorterRuntime(t)

	testFiles := []string{
		"emptyoutput",
		"jsonoutput",
		"multiliner",
		"oneliner",
	}

	for _, testFile := range testFiles {
		r.TestContext.AddTestFile(
			fmt.Sprintf("testdata/outputs/%s.txt", testFile),
			filepath.Join(context.MixinOutputsDir, testFile))
	}

	gotOutputs, err := r.readMixinOutputs()
	require.NoError(t, err)

	for _, testFile := range testFiles {
		if exists, _ := r.Context.FileSystem.Exists(testFile); exists {
			require.Fail(t, fmt.Sprintf("file %s should not exist after reading outputs", testFile))
		}
	}

	wantOutputs := map[string]string{
		"emptyoutput": "",
		"jsonoutput": `{
  "foo": true,
  "things": [
    123,
    "abc",
    false
  ]
}`,
		"multiliner": `FOO

BAR
BAZ`,
		"oneliner": "ABC",
	}

	assert.Equal(t, wantOutputs, gotOutputs)
}

func TestPorterRuntime_ApplyStepOutputsToBundle_None(t *testing.T) {
	r := NewTestPorterRuntime(t)
	m := &manifest.Manifest{Name: "mybun"}
	r.RuntimeManifest = NewRuntimeManifest(r.Context, claim.ActionInstall, m)

	outputs := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	err := r.applyStepOutputsToBundle(outputs)
	assert.NoError(t, err)
}

func TestPorterRuntime_ApplyStepOutputsToBundle_Some_Match(t *testing.T) {
	r := NewTestPorterRuntime(t)
	m := &manifest.Manifest{
		Name: "mybun",
		Outputs: manifest.OutputDefinitions{
			"foo": {
				Name: "foo",
				Schema: definition.Schema{
					Type: "string",
				},
				Sensitive: true,
			},
			"123": {
				Name: "123",
				Schema: definition.Schema{
					Type: "string",
				},
				Sensitive: false,
			},
		},
	}
	r.RuntimeManifest = NewRuntimeManifest(r.Context, claim.ActionInstall, m)

	outputs := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	err := r.applyStepOutputsToBundle(outputs)
	assert.NoError(t, err)

	want := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	for _, outputName := range []string{"foo", "123"} {
		bytes, err := r.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, outputName))
		assert.NoError(t, err)

		assert.Equal(t, want[outputName], string(bytes))
	}
}

func TestPorterRuntime_ApplyStepOutputsToBundle_Some_NoMatch(t *testing.T) {
	r := NewTestPorterRuntime(t)
	m := &manifest.Manifest{
		Name: "mybun",
		Outputs: manifest.OutputDefinitions{
			"bar": {
				Name: "bar",
			},
			"456": {
				Name: "456",
			},
		},
	}
	r.RuntimeManifest = NewRuntimeManifest(r.Context, claim.ActionInstall, m)

	outputs := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	err := r.applyStepOutputsToBundle(outputs)
	assert.NoError(t, err)

	// No outputs declared in the manifest match those in outputs,
	// so no output file is expected to be written
	for _, output := range []string{"foo", "bar", "123", "456"} {
		_, err := r.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, output))
		assert.Error(t, err)
	}
}

func TestPorterRuntime_ApplyStepOutputsToBundle_ApplyTo_True(t *testing.T) {
	r := NewTestPorterRuntime(t)
	m := &manifest.Manifest{
		Name: "mybun",
		Outputs: manifest.OutputDefinitions{
			"foo": {
				Name: "foo",
				ApplyTo: []string{
					"upgrade",
				},
			},
			"123": {
				Name: "123",
				ApplyTo: []string{
					"install",
				},
				Schema: definition.Schema{
					Type: "string",
				},
				Sensitive: false,
			},
		},
	}
	r.RuntimeManifest = NewRuntimeManifest(r.Context, claim.ActionInstall, m)

	outputs := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	err := r.applyStepOutputsToBundle(outputs)
	assert.NoError(t, err)

	// foo output should not exist (applyTo doesn't match)
	_, err = r.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, "foo"))
	assert.Error(t, err)

	// 123 output should exist (applyTo matches)
	bytes, err := r.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, "123"))
	assert.NoError(t, err)

	want := "abc"
	assert.Equal(t, want, string(bytes))
}

func TestPorterRuntime_ApplyUnboundBundleOutputs_File(t *testing.T) {
	const srcPath = "/root/.kube/config"
	const outputName = "kubeconfig"

	testcases := []struct {
		name       string
		shouldBind bool
		def        manifest.OutputDefinition
	}{
		{name: "file with applyto",
			shouldBind: true,
			def: manifest.OutputDefinition{
				Name: outputName,
				ApplyTo: []string{
					"install",
				},
				Schema: definition.Schema{
					Type: "file",
				},
				Path: srcPath,
			},
		},
		{name: "file no applyto",
			shouldBind: true,
			def: manifest.OutputDefinition{
				Name: outputName,
				Schema: definition.Schema{
					Type: "file",
				},
				Path: srcPath,
			},
		},
		{name: "not file",
			shouldBind: true,
			def: manifest.OutputDefinition{
				Name: outputName,
				Schema: definition.Schema{
					Type: "string",
				},
				Path: srcPath,
			},
		},
		{name: "file no path",
			shouldBind: false,
			def: manifest.OutputDefinition{
				Name: outputName,
				Schema: definition.Schema{
					Type: "string",
				},
				Path: "",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewTestPorterRuntime(t)
			m := &manifest.Manifest{
				Name: "mybun",
				Outputs: manifest.OutputDefinitions{
					tc.def.Name: tc.def,
				},
			}
			r.RuntimeManifest = NewRuntimeManifest(r.Context, claim.ActionInstall, m)

			_, err := r.FileSystem.Create(srcPath)
			require.NoError(t, err)

			err = r.applyUnboundBundleOutputs()
			require.NoError(t, err)

			exists, _ := r.FileSystem.Exists("/cnab/app/outputs/" + outputName)
			assert.Equal(t, exists, tc.shouldBind)
		})
	}
}

func TestPorterRuntime_LoadImageMappingFilesNoRelo(t *testing.T) {
	r := NewTestPorterRuntime(t)
	r.TestContext.AddTestFile("testdata/bundle-images.json", "/cnab/bundle.json")
	bun, reloMap, err := r.getImageMappingFiles()
	assert.NoError(t, err)
	assert.Empty(t, reloMap)
	assert.Equal(t, "mysql", bun.Name)
}

func TestLoadImageMappingFilesNoBundle(t *testing.T) {
	r := NewTestPorterRuntime(t)
	r.TestContext.AddTestFile("testdata/relocation-mapping.json", "/cnab/app/relocation-mapping.json")
	_, _, err := r.getImageMappingFiles()
	assert.True(t, os.IsNotExist(errors.Cause(err)), "expected this to fail because bundle.json doesn't exist")
}

func TestLoadImageMappingFilesBadBundle(t *testing.T) {
	r := NewTestPorterRuntime(t)
	r.TestContext.AddTestFile("../porter/testdata/porter.yaml", "/cnab/bundle.json")
	r.TestContext.AddTestFile("testdata/relocation-mapping.json", "/cnab/app/relocation-mapping.json")
	_, _, err := r.getImageMappingFiles()
	assert.EqualError(t, err, "couldn't load runtime bundle.json: invalid character 'a' in literal null (expecting 'u')")
}

func TestLoadImageMappingFilesGoodFiles(t *testing.T) {
	r := NewTestPorterRuntime(t)
	r.TestContext.AddTestFile("testdata/bundle-images.json", "/cnab/bundle.json")
	r.TestContext.AddTestFile("testdata/relocation-mapping.json", "/cnab/app/relocation-mapping.json")

	bun, reloMap, err := r.getImageMappingFiles()
	assert.NoError(t, err)
	assert.NotEmpty(t, reloMap)
	assert.Equal(t, "mysql", bun.Name)
}
