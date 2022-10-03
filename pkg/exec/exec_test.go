package exec

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/exec/builder"
	"get.porter.sh/porter/pkg/test"
	"get.porter.sh/porter/pkg/yaml"
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
	assert.Equal(t, step.EnvironmentVars, map[string]string{"SECRET": "super-secret"})
	require.Len(t, step.Outputs, 1)
	assert.Equal(t, Output{Name: "vms", JsonPath: "$[*].id"}, step.Outputs[0])
}

func TestMixin_ExecuteCommand(t *testing.T) {
	ctx := context.Background()

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
	m.Config.In = bytes.NewReader(b)

	m.Config.Setenv(test.ExpectedCommandEnv, `bash -c echo Hello World`)

	err := m.Execute(ctx, ExecuteOptions{})

	require.NoError(t, err)
}

func TestMixin_ErrorHandling(t *testing.T) {
	testcases := []struct {
		name      string
		handler   builder.IgnoreErrorHandler
		wantError string
	}{
		{name: "legit error", handler: builder.IgnoreErrorHandler{}, wantError: "error running command"},
		{name: "all", handler: builder.IgnoreErrorHandler{All: true}},
		{name: "exit code", handler: builder.IgnoreErrorHandler{ExitCodes: []int{1}}},
		{name: "contains", handler: builder.IgnoreErrorHandler{Output: builder.IgnoreErrorWithOutput{Contains: []string{"already exists"}}}},
		{name: "regex", handler: builder.IgnoreErrorHandler{Output: builder.IgnoreErrorWithOutput{Regex: []string{".* exists"}}}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			step := Step{
				Instruction: Instruction{
					Command:            "bash",
					Arguments:          []string{"-c", "echo Hello World"},
					IgnoreErrorHandler: tc.handler,
				},
			}
			action := Action{
				Name:  "install",
				Steps: []Step{step},
			}
			b, _ := yaml.Marshal(action)

			m := NewTestMixin(t)
			m.Config.In = bytes.NewReader(b)

			m.Config.Setenv(test.ExpectedCommandEnv, `bash -c echo Hello World`)
			m.Config.Setenv(test.ExpectedCommandExitCodeEnv, "1")
			m.Config.Setenv(test.ExpectedCommandErrorEnv, "thing already exists")

			err := m.Execute(ctx, ExecuteOptions{})
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantError)
			}
		})
	}
}

func TestMixin_Install(t *testing.T) {
	ctx := context.Background()
	h := NewTestMixin(t)
	h.TestConfig.TestContext.AddTestDirectory("testdata", "testdata")

	action, err := h.loadAction(ctx, "testdata/install-input.yaml")
	require.NoError(t, err)

	assert.Len(t, action.Steps, 1)
	step := action.Steps[0]
	assert.Equal(t, "gcloud", step.Instruction.Command)
	require.Len(t, step.Outputs, 1)
	assert.Equal(t, "$[*].id", step.Outputs[0].JsonPath)
}

func TestMixin_Upgrade(t *testing.T) {
	ctx := context.Background()
	h := NewTestMixin(t)
	h.TestConfig.TestContext.AddTestDirectory("testdata", "testdata")

	action, err := h.loadAction(ctx, "testdata/upgrade-input.yaml")
	require.NoError(t, err)

	assert.Len(t, action.Steps, 1)
	step := action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
	require.Len(t, step.Outputs, 1)
	assert.Equal(t, "config/kube.yaml", step.Outputs[0].FilePath)
}

func TestMixin_CustomAction(t *testing.T) {
	ctx := context.Background()
	h := NewTestMixin(t)
	h.TestConfig.TestContext.AddTestDirectory("testdata", "testdata")

	action, err := h.loadAction(ctx, "testdata/invoke-input.yaml")
	require.NoError(t, err)

	assert.Len(t, action.Steps, 1)
	step := action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
	require.Len(t, step.Outputs, 1)
	assert.Equal(t, "Hello (.*)", step.Outputs[0].Regex)
}

func TestMixin_Uninstall(t *testing.T) {
	ctx := context.Background()
	h := NewTestMixin(t)
	h.TestConfig.TestContext.AddTestDirectory("testdata", "testdata")

	action, err := h.loadAction(ctx, "testdata/uninstall-input.yaml")
	require.NoError(t, err)

	assert.Len(t, action.Steps, 1)
	step := action.Steps[0]
	assert.Equal(t, "bash", step.Instruction.Command)
}

func TestMixin_SuffixArgs(t *testing.T) {
	ctx := context.Background()
	b, err := ioutil.ReadFile("testdata/suffix-args-input.yaml")
	require.NoError(t, err, "ReadFile failed")

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err, "Unmarshal failed")

	m := NewTestMixin(t)
	m.Config.In = bytes.NewReader(b)

	m.Config.Setenv(test.ExpectedCommandEnv, `docker build --tag getporter/porter-hello:latest .`)

	err = m.Execute(ctx, ExecuteOptions{})
	require.NoError(t, err)
}
