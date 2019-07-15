package exec

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestAction_UnmarshalYAML(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/install_input.yaml")
	require.NoError(t, err)

	action := Action{}
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)

	assert.Len(t, action.Steps, 1)
	step := action.Steps[0]
	assert.Equal(t, "bash", step.Command)
	assert.Equal(t, "Install Hello World", step.Description)
	assert.Len(t, step.Arguments, 2)
	assert.Equal(t, "-c", step.Arguments[0])
	assert.Equal(t, "echo Hello World", step.Arguments[1])
}

func TestMixin_ExecuteCommand(t *testing.T) {
	os.Setenv(test.ExpectedCommandEnv, `bash -c echo Hello World`)
	defer os.Unsetenv(test.ExpectedCommandEnv)

	step := Step{
		Instruction: Instruction{
			Command:   "bash",
			Arguments: []string{"-c", "echo Hello World"},
		},
	}
	action := Action{
		Steps: []Step{step},
	}
	b, _ := yaml.Marshal(action)

	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)

	err := h.ExecuteCommand(ExecuteCommandOptions{})

	require.NoError(t, err)
}

func TestMixin_Install(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	err := h.loadAction("testdata/install_input.yaml")
	require.NoError(t, err)

	assert.Len(t, h.Mixin.Action.Steps, 1)
	step := h.Mixin.Action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
}

func TestMixin_Upgrade(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	err := h.loadAction("testdata/upgrade_input.yaml")
	require.NoError(t, err)

	assert.Len(t, h.Mixin.Action.Steps, 1)
	step := h.Mixin.Action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
}

func TestMixin_CustomAction(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	err := h.loadAction("testdata/status_input.yaml")
	require.NoError(t, err)

	assert.Len(t, h.Mixin.Action.Steps, 1)
	step := h.Mixin.Action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
}

func TestMixin_Uninstall(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	err := h.loadAction("testdata/uninstall_input.yaml")
	require.NoError(t, err)

	assert.Len(t, h.Mixin.Action.Steps, 1)
	step := h.Mixin.Action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
}
