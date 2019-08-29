package builder

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStep struct {
	Command   string
	Arguments []string
	Flags     Flags
	Outputs   []Output
}

func (s TestStep) GetCommand() string {
	return s.Command
}

func (s TestStep) GetArguments() []string {
	return s.Arguments
}

func (s TestStep) GetFlags() Flags {
	return s.Flags
}

func (s TestStep) GetOutputs() []Output {
	return s.Outputs
}

type TestJsonPathOutput struct {
	Name     string
	JsonPath string
}

func (o TestJsonPathOutput) GetName() string {
	return o.Name
}

func (o TestJsonPathOutput) GetJsonPath() string {
	return o.JsonPath
}

func TestJsonPathOutputs(t *testing.T) {
	c := context.NewTestContext(t)

	step := TestStep{
		Outputs: []Output{
			TestJsonPathOutput{Name: "ids", JsonPath: "$[*].id"},
			TestJsonPathOutput{Name: "names", JsonPath: "$[*].name"},
		},
	}
	stdout, err := ioutil.ReadFile("testdata/install-output.json")
	require.NoError(t, err, "could not read testdata")

	err = ProcessJsonPathOutputs(c.Context, step, string(stdout))
	require.NoError(t, err, "ProcessJsonPathOutputs should not return an error")

	f := filepath.Join(context.MixinOutputsDir, "ids")
	gotOutput, err := c.FileSystem.ReadFile(f)
	require.NoError(t, err, "could not read output file %s", f)

	wantOutput := `["1085517466897181794"]`
	assert.Equal(t, wantOutput, string(gotOutput))

	f = filepath.Join(context.MixinOutputsDir, "names")
	gotOutput, err = c.FileSystem.ReadFile(f)
	require.NoError(t, err, "could not read output file %s", f)

	wantOutput = `["porter-test"]`
	assert.Equal(t, wantOutput, string(gotOutput))
}

func TestJsonPathOutputs_NoOutput(t *testing.T) {
	c := context.NewTestContext(t)

	step := TestStep{
		Outputs: []Output{
			TestJsonPathOutput{Name: "ids", JsonPath: "$[*].id"},
		},
	}

	err := ProcessJsonPathOutputs(c.Context, step, "")
	require.NoError(t, err, "ProcessJsonPathOutputs should not return an error when the output buffer is empty")
}
