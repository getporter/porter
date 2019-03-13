package exec

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestAction_UnmarshalYAML(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/exec_input.yaml")
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
