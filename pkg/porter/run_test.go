package porter

import (
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

	wantOutputs := []string{
		"FOO=BAR",
		"BAZ=QUX",
		"A=B",
	}
	assert.Equal(t, wantOutputs, gotOutputs)
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
