package porter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/config"
	yaml "gopkg.in/yaml.v2"

	"github.com/deislabs/porter/pkg/mixin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_readOutputs(t *testing.T) {
	p := NewTestPorter(t)

	p.TestConfig.TestContext.AddTestFile("testdata/outputs1.txt", filepath.Join(mixin.OutputsDir, "myoutput1"))
	p.TestConfig.TestContext.AddTestFile("testdata/outputs2.txt", filepath.Join(mixin.OutputsDir, "myoutput2"))

	gotOutputs, err := p.readOutputs()
	require.NoError(t, err)

	for _, file := range []string{filepath.Join(mixin.OutputsDir, "myoutput1"), filepath.Join(mixin.OutputsDir, "myoutput2")} {
		if exists, _ := p.FileSystem.Exists(file); exists {
			require.Fail(t, fmt.Sprintf("file %s should not exist after reading outputs", file))
		}
	}

	wantOutputs := []string{
		"FOO=BAR",
		"BAZ=QUX",
		"A=B",
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
	s := &config.Step{
		Data: map[string]interface{}{
			"exec": map[string]interface{}{
				"command": "echo hi",
			},
		},
	}

	input := &ActionInput{
		action: config.ActionInstall,
		Steps:  []*config.Step{s},
	}

	b, err := yaml.Marshal(input)
	require.NoError(t, err)
	wantYaml := "install:\n- exec:\n    command: echo hi\n"
	assert.Equal(t, wantYaml, string(b))
}

func TestApplyBundleOutputs_None(t *testing.T) {
	p := NewTestPorter(t)
	p.Manifest = &config.Manifest{
		Name: "mybun",
	}
	opts := NewRunOptions(p.Config)

	outputs := []string{"foo=bar", "123=abc"}

	err := p.ApplyBundleOutputs(opts, outputs)
	assert.NoError(t, err)
}

func TestApplyBundleOutputs_Some_Match(t *testing.T) {
	p := NewTestPorter(t)
	p.Manifest = &config.Manifest{
		Name: "mybun",
		Outputs: []config.OutputDefinition{
			{
				Name: "foo",
			},
			{
				Name: "123",
			},
		},
	}
	opts := NewRunOptions(p.Config)

	outputs := []string{"foo=bar", "123=abc"}

	err := p.ApplyBundleOutputs(opts, outputs)
	assert.NoError(t, err)

	bytes, err := p.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, "foo"))
	assert.NoError(t, err)
	assert.Equal(t, outputs[0], string(bytes))

	bytes, err = p.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, "123"))
	assert.NoError(t, err)
	assert.Equal(t, outputs[1], string(bytes))
}

func TestApplyBundleOutputs_Some_NoMatch(t *testing.T) {
	p := NewTestPorter(t)
	p.Manifest = &config.Manifest{
		Name: "mybun",
		Outputs: []config.OutputDefinition{
			{
				Name: "bar",
			},
			{
				Name: "456",
			},
		},
	}
	opts := NewRunOptions(p.Config)

	outputs := []string{"foo=bar", "123=abc"}

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
	p.Manifest = &config.Manifest{
		Name: "mybun",
		Outputs: []config.OutputDefinition{
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
			},
		},
	}
	opts := NewRunOptions(p.Config)
	opts.Action = "install"

	outputs := []string{"foo=bar", "123=abc"}

	err := p.ApplyBundleOutputs(opts, outputs)
	assert.NoError(t, err)

	// foo output should not exist (applyTo doesn't match)
	_, err = p.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, "foo"))
	assert.Error(t, err)

	// 123 output should exist (applyTo matches)
	bytes, err := p.FileSystem.ReadFile(filepath.Join(config.BundleOutputsDir, "123"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("123=abc"), bytes)
}
