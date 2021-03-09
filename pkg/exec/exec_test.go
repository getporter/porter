package exec

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/exec/builder"

	"get.porter.sh/porter/pkg/test"

	yaml "get.porter.sh/porter/pkg/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	m := NewTestMixin(t)
	m.In = bytes.NewReader(b)

	m.Setenv(test.ExpectedCommandEnv, `bash -c echo Hello World`)

	err := m.Execute(ExecuteOptions{})

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

func TestMixin_SuffixArgs(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/suffix-args-input.yaml")
	require.NoError(t, err, "ReadFile failed")

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err, "Unmarshal failed")

	m := NewTestMixin(t)
	m.In = bytes.NewReader(b)

	m.Setenv(test.ExpectedCommandEnv, `docker build --tag getporter/porter-hello:latest .`)

	err = m.Execute(ExecuteOptions{})
	require.NoError(t, err)
}
