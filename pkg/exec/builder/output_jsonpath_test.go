package builder

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStep struct {
	Outputs []Output
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
	output, err := ioutil.ReadFile("testdata/install-output.json")
	require.NoError(t, err, "could not read testdata")

	err = ProcessJsonPathOutputs(c.Context, step, bytes.NewBuffer(output))
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

	err := ProcessJsonPathOutputs(c.Context, step, bytes.NewBufferString(""))
	require.NoError(t, err, "ProcessJsonPathOutputs should not return an error when the output buffer is empty")
}
