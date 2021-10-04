package builder

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStep struct {
	Command          string
	Arguments        []string
	Flags            Flags
	Outputs          []Output
	WorkingDirectory string
}

func (s TestStep) GetCommand() string {
	return s.Command
}

func (s TestStep) GetArguments() []string {
	return s.Arguments
}

func (s TestStep) GetWorkingDir() string {
	return s.WorkingDirectory
}

func (s TestStep) GetFlags() Flags {
	return s.Flags
}

func (s TestStep) GetDashes() Dashes {
	return DefaultFlagDashes
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
	testcases := []struct {
		name       string
		jsonPath   string
		wantOutput string
	}{
		{"array", "$[*].id", `["1085517466897181794"]`},
		{"object", "$[0].tags", `{"fingerprint":"42WmSpB8rSM="}`},
		{"integer", "$[0].index", `0`},
		{"big integer", "$[0]._id", `123123123`},
		{"exponential notation", "$[0]._bigId", `1.23123123e+08`},
		{"boolean", "$[0].deletionProtection", `false`},
		{"string", "$[0].cpuPlatform", `Intel Haswell`},
	}

	stdout, err := ioutil.ReadFile("testdata/install-output.json")
	require.NoError(t, err, "could not read testdata")

	for _, tc := range testcases {

		t.Run(tc.name, func(t *testing.T) {
			c := context.NewTestContext(t)

			step := TestStep{
				Outputs: []Output{
					TestJsonPathOutput{Name: tc.name, JsonPath: tc.jsonPath},
				},
			}

			err = ProcessJsonPathOutputs(c.Context, step, string(stdout))
			require.NoError(t, err, "ProcessJsonPathOutputs should not return an error")

			f := filepath.Join(context.MixinOutputsDir, tc.name)
			gotOutput, err := c.FileSystem.ReadFile(f)
			require.NoError(t, err, "could not read output file %s", f)

			wantOutput := tc.wantOutput
			assert.Equal(t, wantOutput, string(gotOutput))
		})
	}
}

func TestJsonPathOutputs_DebugPrintsDocument(t *testing.T) {
	c := context.NewTestContext(t)
	c.Debug = true
	step := TestStep{
		Outputs: []Output{
			TestJsonPathOutput{Name: "ids", JsonPath: "$[*].id"},
		},
	}

	document := `[{"id": "abc123"}]`
	err := ProcessJsonPathOutputs(c.Context, step, document)
	require.NoError(t, err)
	wantDebugLine := `Processing jsonpath output ids using query $[*].id against document
[{"id": "abc123"}]
`
	assert.Contains(t, c.GetError(), wantDebugLine, "Debug mode should print the full document and query being captured")
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
