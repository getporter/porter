package exec

import (
	"bytes"
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/test"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// sad hack: not sure how to make a common test main for all my subpackages
func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}

func TestMixin_Install(t *testing.T) {
	os.Setenv(test.ExpectedCommandEnv, `bash -c echo Hello World`)
	defer os.Unsetenv(test.ExpectedCommandEnv)

	step := Step{
		Instruction: Instruction{
			Command:   "bash",
			Arguments: []string{"-c", "echo Hello World"},
		},
	}
	b, _ := yaml.Marshal(step)

	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)

	err := h.Install("")

	require.NoError(t, err)
}
