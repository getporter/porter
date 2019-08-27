package exec

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/exec/builder"

	"github.com/deislabs/porter/pkg/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestAction_UnmarshalYAML(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/install-input.yaml")
	require.NoError(t, err)

	action := Action{}
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)

	require.Len(t, action.Steps, 1)
	step := action.Steps[0]
	assert.Equal(t, "bash", step.Command)
	assert.Equal(t, "Install Hello World", step.Description)
	assert.Len(t, step.Flags, 1)
	assert.Equal(t, builder.NewFlag("c", "echo Hello World"), step.Flags[0])
	assert.Len(t, step.Arguments, 0)
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

	fmt.Println(string(b))
	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)

	err := h.ExecuteCommand(ExecuteCommandOptions{})

	require.NoError(t, err)
}

func TestMixin_Install(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	err := h.loadAction("testdata/install-input.yaml")
	require.NoError(t, err)

	assert.Len(t, h.Mixin.Action.Steps, 1)
	step := h.Mixin.Action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
}

func TestMixin_Upgrade(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	err := h.loadAction("testdata/upgrade-input.yaml")
	require.NoError(t, err)

	assert.Len(t, h.Mixin.Action.Steps, 1)
	step := h.Mixin.Action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
}

func TestMixin_CustomAction(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	err := h.loadAction("testdata/invoke-input.yaml")
	require.NoError(t, err)

	assert.Len(t, h.Mixin.Action.Steps, 1)
	step := h.Mixin.Action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
}

func TestMixin_Uninstall(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	err := h.loadAction("testdata/uninstall-input.yaml")
	require.NoError(t, err)

	assert.Len(t, h.Mixin.Action.Steps, 1)
	step := h.Mixin.Action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
}
