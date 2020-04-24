package exec

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/exec/builder"

	"get.porter.sh/porter/pkg/test"

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
	assert.Equal(t, "gcloud", step.Command)
	assert.Equal(t, "Install a VM and collect its ID", step.Description)
	require.Len(t, step.Arguments, 3)
	assert.Len(t, step.Flags, 3)
	sort.Sort(step.Flags) // This returns the flags in ascending order
	assert.Equal(t, builder.NewFlag("machine-type", "f1-micro"), step.Flags[0])
	assert.Equal(t, builder.NewFlag("project", "porterci"), step.Flags[1])
	assert.Equal(t, builder.NewFlag("zone", "us-central1-a"), step.Flags[2])
	require.Len(t, step.Outputs, 1)
	assert.Equal(t, Output{Name: "vms", JsonPath: "$[*].id"}, step.Outputs[0])
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
		Name:  "install",
		Steps: []Step{step},
	}
	b, _ := yaml.Marshal(action)

	fmt.Println(string(b))
	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)

	err := h.Execute(ExecuteOptions{})

	require.NoError(t, err)
}

func TestMixin_Install(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	action, err := h.loadAction("testdata/install-input.yaml")
	require.NoError(t, err)

	assert.Len(t, action.Steps, 1)
	step := action.Steps[0]
	assert.Equal(t, "gcloud", step.Instruction.Command)
	require.Len(t, step.Outputs, 1)
	assert.Equal(t, "$[*].id", step.Outputs[0].JsonPath)
}

func TestMixin_Upgrade(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	action, err := h.loadAction("testdata/upgrade-input.yaml")
	require.NoError(t, err)

	assert.Len(t, action.Steps, 1)
	step := action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
	require.Len(t, step.Outputs, 1)
	assert.Equal(t, "config/kube.yaml", step.Outputs[0].FilePath)
}

func TestMixin_CustomAction(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	action, err := h.loadAction("testdata/invoke-input.yaml")
	require.NoError(t, err)

	assert.Len(t, action.Steps, 1)
	step := action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
	require.Len(t, step.Outputs, 1)
	assert.Equal(t, "Hello (.*)", step.Outputs[0].Regex)
}

func TestMixin_Uninstall(t *testing.T) {
	h := NewTestMixin(t)
	h.TestContext.AddTestDirectory("testdata", "testdata")

	action, err := h.loadAction("testdata/uninstall-input.yaml")
	require.NoError(t, err)

	assert.Len(t, action.Steps, 1)
	step := action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
}
