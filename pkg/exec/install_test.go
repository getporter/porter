package exec

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deislabs/porter/pkg/test"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestMixin_Install(t *testing.T) {
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

	err := h.Install("")

	require.NoError(t, err)
}

func TestMixin_LoadInstructionFromFile(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	err := h.loadAction("testdata/exec_input.yaml")
	require.NoError(t, err)

	assert.Len(t, h.Mixin.Action.Steps, 1)
	step := h.Mixin.Action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
}
