package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg"
	porterruntime "get.porter.sh/porter/pkg/runtime"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestAction struct {
	Steps []TestStep
}

func (a TestAction) GetSteps() []ExecutableStep {
	steps := make([]ExecutableStep, len(a.Steps))
	for i := range a.Steps {
		steps[i] = a.Steps[i]
	}
	return steps
}

func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}

func TestExecuteSingleStepAction(t *testing.T) {
	ctx := context.Background()
	c := porterruntime.NewTestRuntimeConfig(t)

	err := c.FileSystem.WriteFile("config.txt", []byte("abc123"), pkg.FileModeWritable)
	require.NoError(t, err)

	a := TestAction{
		Steps: []TestStep{
			{
				Command: "foo",
				Outputs: []Output{
					TestFileOutput{Name: "file", FilePath: "config.txt"},
					TestJsonPathOutput{Name: "jsonpath", JsonPath: "$.*"},
					TestRegexOutput{Name: "regex", Regex: "(.*)"},
				}},
		},
	}

	c.Setenv(test.ExpectedCommandEnv, "foo")

	_, err = ExecuteSingleStepAction(ctx, c.RuntimeConfig, a)
	require.NoError(t, err, "ExecuteSingleStepAction should not have returned an error")

	exists, _ := c.FileSystem.Exists("/cnab/app/porter/outputs/file")
	assert.True(t, exists, "file output was not evaluated")

	exists, _ = c.FileSystem.Exists("/cnab/app/porter/outputs/regex")
	assert.True(t, exists, "regex output was not evaluated")

	exists, _ = c.FileSystem.Exists("/cnab/app/porter/outputs/jsonpath")
	assert.True(t, exists, "jsonpath output was not evaluated")
}

func Test_splitCommand(t *testing.T) {
	t.Run("split space", func(t *testing.T) {
		result := splitCommand([]string{"cmd", "--myarg", "val1 val2"})
		assert.Equal(t, []string{"cmd", "--myarg", "val1", "val2"}, result, "strings not enclosed should be split apart")
	})

	t.Run("split tab", func(t *testing.T) {
		result := splitCommand([]string{"cmd", "--myarg", "val1\tval2"})
		assert.Equal(t, []string{"cmd", "--myarg", "val1", "val2"}, result, "strings not enclosed should be split apart")
	})

	t.Run("split newline", func(t *testing.T) {
		result := splitCommand([]string{"cmd", "--myarg", "val1\nval2"})
		assert.Equal(t, []string{"cmd", "--myarg", "val1", "val2"}, result, "strings not enclosed should be split apart")
	})

	t.Run("keep double quoted whitespace", func(t *testing.T) {
		result := splitCommand([]string{"cmd", "--myarg", `"val1 val2" val3`})
		assert.Equal(t, []string{"cmd", "--myarg", "val1 val2", "val3"}, result, "strings in the enclosing quotes should be grouped together")
	})

	t.Run("embedded single quote", func(t *testing.T) {
		result := splitCommand([]string{"cmd", "--myarg", `"Patty O'Brien" true`})
		assert.Equal(t, []string{"cmd", "--myarg", "Patty O'Brien", "true"}, result, "single quotes should be included in the enclosing quotes")
	})

	t.Run("escaped double quotes", func(t *testing.T) {
		result := splitCommand([]string{"c", `"echo { \"test\": \"myvalue\" }"`})
		assert.Equal(t, []string{"c", `echo { \"test\": \"myvalue\" }`}, result, "escaped double quotes should be included in the enclosing quotes")
	})

	t.Run("escaped single quotes", func(t *testing.T) {
		result := splitCommand([]string{"c", `"echo $'I\'m a linux admin.'"`})
		assert.Equal(t, []string{"c", `echo $'I\'m a linux admin.'`}, result, "escaped single quotes should be included in the enclosing quotes")
	})

	t.Run("unmatched double quote", func(t *testing.T) {
		result := splitCommand([]string{"cmd", "--myarg", `"Patty O"Brien" true`})
		assert.Equal(t, []string{"cmd", "--myarg", `"Patty O"Brien" true`}, result, "unmatched double quotes should cause the grouping to fail")
	})

	t.Run("unmatched single quote", func(t *testing.T) {
		result := splitCommand([]string{"cmd", "--myarg", `'Patty O'Brien' true`})
		assert.Equal(t, []string{"cmd", "--myarg", `'Patty O'Brien' true`}, result, "unmatched single quotes should cause the grouping to fail")
	})
}

var _ HasOrderedArguments = TestOrderedStep{}

type TestOrderedStep struct {
	TestStep
	SuffixArguments []string
}

func (s TestOrderedStep) GetSuffixArguments() []string {
	return s.SuffixArguments
}

func TestExecuteStep_HasOrderedArguments(t *testing.T) {
	ctx := context.Background()
	c := porterruntime.NewTestRuntimeConfig(t)
	step := TestOrderedStep{
		TestStep: TestStep{
			Command:   "docker",
			Arguments: []string{"build"},
			Flags: []Flag{
				NewFlag("t", "getporter/porter-hello:latest"),
			},
		},
		SuffixArguments: []string{"."},
	}

	c.Setenv(test.ExpectedCommandEnv, "docker build -t getporter/porter-hello:latest .")

	_, err := ExecuteStep(ctx, c.RuntimeConfig, step)
	require.NoError(t, err, "ExecuteStep failed")
}

func TestExecuteStep_SpecifiesCustomWorkingDirectory(t *testing.T) {
	ctx := context.Background()
	c := porterruntime.NewTestRuntimeConfig(t)
	c.TestContext.UseFilesystem()
	wd, _ := filepath.EvalSymlinks(os.TempDir())

	step := TestOrderedStep{
		TestStep: TestStep{
			Command:          "pwd",
			Arguments:        []string{},
			WorkingDirectory: wd,
		},
		SuffixArguments: []string{},
	}

	_, err := ExecuteStep(ctx, c.RuntimeConfig, step)
	assert.Equal(t, fmt.Sprintln(wd), c.TestContext.GetOutput())
	require.NoError(t, err, "Execute Step failed")
}

func (s TestOrderedStep) GetEnvironmentVars() map[string]string {
	return s.EnvironmentVars
}

func TestExecuteStep_WithEnvironmentVars(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}

	ctx := context.Background()
	c := porterruntime.NewTestRuntimeConfig(t)
	c.TestContext.UseFilesystem()
	step := TestOrderedStep{
		TestStep: TestStep{
			Command:         "env",
			EnvironmentVars: map[string]string{"SOME_VAR_123": "foo"},
		},
	}

	c.Setenv(test.ExpectedCommandEnv, "env")

	_, err := ExecuteStep(ctx, c.RuntimeConfig, step)
	require.NoError(t, err, "Execute Step failed")
	containsEnv := strings.Contains(c.TestContext.GetOutput(), "SOME_VAR_123=foo")
	// use assert.True rather than assert.Contains so that the env vars are not all sent to the test output. There might
	// be sensitive stuff in there.
	assert.True(t, containsEnv, "Env did not contain the key/value we expected.")
}
