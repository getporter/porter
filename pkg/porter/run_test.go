package porter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/manifest"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestPorter_Run(t *testing.T) {
	p := NewTestPorter(t)

	// Mock the mixin test runner and verify that we are calling runtime mixins, e.g. exec-runtime and not exec
	mp := p.Mixins.(*mixin.TestMixinProvider)
	mp.RunAssertions = append(mp.RunAssertions, func(mixinCxt *context.Context, mixinName string, commandOpts mixin.CommandOptions) {
		assert.Equal(t, "exec", mixinName, "expected to call the exec mixin")
		assert.True(t, commandOpts.Runtime, "the mixin command should be executed in runtime mode")
		assert.Equal(t, "install", commandOpts.Command, "should have executed the mixin's install command")
	})
	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/cnab/bundle.json")
	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")

	opts := NewRunOptions(p.Config)
	opts.Action = string(manifest.ActionInstall)
	opts.File = "porter.yaml"

	err := opts.Validate()
	require.NoError(t, err, "could not validate run options")

	err = p.Run(opts)
	assert.NoError(t, err, "run failed")
}

func TestPorter_readMixinOutputs(t *testing.T) {
	p := NewTestPorter(t)

	testFiles := []string{
		"emptyoutput",
		"jsonoutput",
		"multiliner",
		"oneliner",
	}

	for _, testFile := range testFiles {
		p.TestConfig.TestContext.AddTestFile(
			fmt.Sprintf("testdata/outputs/%s.txt", testFile),
			filepath.Join(context.MixinOutputsDir, testFile))
	}

	gotOutputs, err := p.readMixinOutputs()
	require.NoError(t, err)

	for _, testFile := range testFiles {
		if exists, _ := p.FileSystem.Exists(testFile); exists {
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

func TestPorter_defaultDebugToOff(t *testing.T) {
	p := New() // Don't use the test porter, it has debug on by default
	opts := NewRunOptions(p.Config)

	err := opts.defaultDebug()
	require.NoError(t, err)
	assert.False(t, p.Config.Debug)
}

func TestPorter_defaultDebugUsesEnvVar(t *testing.T) {
	os.Setenv(config.EnvDEBUG, "true")
	defer os.Unsetenv(config.EnvDEBUG)

	p := New() // Don't use the test porter, it has debug on by default
	opts := NewRunOptions(p.Config)

	err := opts.defaultDebug()
	require.NoError(t, err)

	assert.True(t, p.Config.Debug)
}

func TestActionInput_MarshalYAML(t *testing.T) {
	s := &manifest.Step{
		Data: map[string]interface{}{
			"exec": map[string]interface{}{
				"command": "echo hi",
			},
		},
	}

	input := &ActionInput{
		action: manifest.ActionInstall,
		Steps:  []*manifest.Step{s},
	}

	b, err := yaml.Marshal(input)
	require.NoError(t, err)
	wantYaml := "install:\n- exec:\n    command: echo hi\n"
	assert.Equal(t, wantYaml, string(b))
}

func TestApplyBundleOutputs_None(t *testing.T) {
	p := NewTestPorter(t)
	p.Manifest = &manifest.Manifest{
		Name: "mybun",
	}
	opts := NewRunOptions(p.Config)

	outputs := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	err := p.ApplyBundleOutputs(opts, outputs)
	assert.NoError(t, err)
}

func TestApplyBundleOutputs_Some_Match(t *testing.T) {
	p := NewTestPorter(t)
	p.Manifest = &manifest.Manifest{
		Name: "mybun",
		Outputs: []manifest.OutputDefinition{
			{
				Name: "foo",
				Schema: definition.Schema{
					Type: "string",
				},
				Sensitive: true,
			},
			{
				Name: "123",
				Schema: definition.Schema{
					Type: "string",
				},
				Sensitive: false,
			},
		},
	}
	opts := NewRunOptions(p.Config)

	outputs := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	err := p.ApplyBundleOutputs(opts, outputs)
	assert.NoError(t, err)

	want := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	for _, outputName := range []string{"foo", "123"} {
		bytes, err := p.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, outputName))
		assert.NoError(t, err)

		assert.Equal(t, want[outputName], string(bytes))
	}
}

func TestApplyBundleOutputs_Some_NoMatch(t *testing.T) {
	p := NewTestPorter(t)
	p.Manifest = &manifest.Manifest{
		Name: "mybun",
		Outputs: []manifest.OutputDefinition{
			{
				Name: "bar",
			},
			{
				Name: "456",
			},
		},
	}
	opts := NewRunOptions(p.Config)

	outputs := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	err := p.ApplyBundleOutputs(opts, outputs)
	assert.NoError(t, err)

	// No outputs declared in the manifest match those in outputs,
	// so no output file is expected to be written
	for _, output := range []string{"foo", "bar", "123", "456"} {
		_, err := p.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, output))
		assert.Error(t, err)
	}
}

func TestApplyBundleOutputs_ApplyTo_True(t *testing.T) {
	p := NewTestPorter(t)
	p.Manifest = &manifest.Manifest{
		Name: "mybun",
		Outputs: []manifest.OutputDefinition{
			{
				Name: "foo",
				ApplyTo: []string{
					"upgrade",
				},
			},
			{
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
	opts := NewRunOptions(p.Config)
	opts.Action = "install"

	outputs := map[string]string{
		"foo": "bar",
		"123": "abc",
	}

	err := p.ApplyBundleOutputs(opts, outputs)
	assert.NoError(t, err)

	// foo output should not exist (applyTo doesn't match)
	_, err = p.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, "foo"))
	assert.Error(t, err)

	// 123 output should exist (applyTo matches)
	bytes, err := p.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, "123"))
	assert.NoError(t, err)

	want := "abc"
	assert.Equal(t, want, string(bytes))
}

func TestLoadImageMappingFilesNoRelo(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.TestContext.AddTestFile("testdata/bundle-images.json", "/cnab/bundle.json")
	bun, reloMap, err := p.getImageMappingFiles()
	assert.NoError(t, err)
	assert.Empty(t, reloMap)
	assert.Equal(t, "mysql", bun.Name)
}

func TestLoadImageMappingFilesNoBundle(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.TestContext.AddTestFile("testdata/relocation-mapping.json", "/cnab/app/relocation-mapping.json")
	_, _, err := p.getImageMappingFiles()
	assert.EqualError(t, err, "couldn't read runtime bundle.json: open /cnab/bundle.json: file does not exist")
}

func TestLoadImageMappingFilesBadBundle(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "/cnab/bundle.json")
	p.TestConfig.TestContext.AddTestFile("testdata/relocation-mapping.json", "/cnab/app/relocation-mapping.json")
	_, _, err := p.getImageMappingFiles()
	assert.EqualError(t, err, "couldn't load runtime bundle.json: invalid character 'a' in literal null (expecting 'u')")
}

func TestLoadImageMappingFilesGoodFiles(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.TestContext.AddTestFile("testdata/bundle-images.json", "/cnab/bundle.json")
	p.TestConfig.TestContext.AddTestFile("testdata/relocation-mapping.json", "/cnab/app/relocation-mapping.json")

	bun, reloMap, err := p.getImageMappingFiles()
	assert.NoError(t, err)
	assert.NotEmpty(t, reloMap)
	assert.Equal(t, "mysql", bun.Name)
}
