package runtime

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/portercontext"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
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
			filepath.Join(portercontext.MixinOutputsDir, testFile))
	}

	gotOutputs, err := r.readMixinOutputs()
	require.NoError(t, err)

	for _, testFile := range testFiles {
		if exists, _ := r.config.FileSystem.Exists(testFile); exists {
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
	r.RuntimeManifest = r.NewRuntimeManifest(cnab.ActionInstall, m)

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
	r.RuntimeManifest = r.NewRuntimeManifest(cnab.ActionInstall, m)

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
		bytes, err := r.config.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, outputName))
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
	r.RuntimeManifest = r.NewRuntimeManifest(cnab.ActionInstall, m)

	outputs := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	err := r.applyStepOutputsToBundle(outputs)
	assert.NoError(t, err)

	// No outputs declared in the manifest match those in outputs,
	// so no output file is expected to be written
	for _, output := range []string{"foo", "bar", "123", "456"} {
		_, err := r.config.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, output))
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
	r.RuntimeManifest = r.NewRuntimeManifest(cnab.ActionInstall, m)

	outputs := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	err := r.applyStepOutputsToBundle(outputs)
	assert.NoError(t, err)

	// foo output should not exist (applyTo doesn't match)
	_, err = r.config.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, "foo"))
	assert.Error(t, err)

	// 123 output should exist (applyTo matches)
	bytes, err := r.config.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, "123"))
	assert.NoError(t, err)

	want := "abc"
	assert.Equal(t, want, string(bytes))
}

func TestRuntimeManifest_ApplyUnboundBundleOutputs_File(t *testing.T) {
	const srcPath = "/home/nonroot/.kube/config"
	const outputName = "kubeconfig"

	ctx := context.Background()
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
			c := portercontext.NewTestContext(t)
			m := &manifest.Manifest{
				Name: "mybun",
				Outputs: manifest.OutputDefinitions{
					tc.def.Name: tc.def,
				},
			}
			cfg := NewConfigFor(c.Context)
			rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
			rm.bundle = cnab.NewBundle(bundle.Bundle{
				Definitions: map[string]*definition.Schema{
					tc.def.Name: &tc.def.Schema,
				},
				Outputs: map[string]bundle.Output{
					tc.def.Name: {
						Definition: tc.def.Name,
						Path:       tc.def.Path,
					},
				},
			})

			_, err := rm.config.FileSystem.Create(srcPath)
			require.NoError(t, err)

			err = rm.applyUnboundBundleOutputs(ctx)
			require.NoError(t, err)

			exists, _ := rm.config.FileSystem.Exists("/cnab/app/outputs/" + outputName)
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read bundle")
}

func TestLoadImageMappingFilesBadBundle(t *testing.T) {
	r := NewTestPorterRuntime(t)
	r.TestContext.AddTestFileFromRoot("pkg/porter/testdata/porter.yaml", "/cnab/bundle.json")
	r.TestContext.AddTestFile("testdata/relocation-mapping.json", "/cnab/app/relocation-mapping.json")
	_, _, err := r.getImageMappingFiles()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot load bundle")
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
