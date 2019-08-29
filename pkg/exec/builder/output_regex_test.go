package builder

import (
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestRegexOutput struct {
	Name  string
	Regex string
}

func (o TestRegexOutput) GetName() string {
	return o.Name
}

func (o TestRegexOutput) GetRegex() string {
	return o.Regex
}

func TestTestRegexOutputs(t *testing.T) {
	c := context.NewTestContext(t)

	step := TestStep{
		Outputs: []Output{
			TestRegexOutput{Name: "failed-test", Regex: `--- FAIL: (.*) \(.*\)`},
		},
	}

	stdout := `--- FAIL: TestMixin_Install (0.00s)
stuff
things
--- FAIL: TestMixin_Upgrade (0.00s)
more
logs`
	err := ProcessRegexOutputs(c.Context, step, stdout)
	require.NoError(t, err, "ProcessRegexOutputs should not return an error")

	f := filepath.Join(context.MixinOutputsDir, "failed-test")
	gotOutput, err := c.FileSystem.ReadFile(f)
	require.NoError(t, err, "could not read output file %s", f)

	wantOutput := `TestMixin_Install
TestMixin_Upgrade`

	assert.Equal(t, wantOutput, string(gotOutput))
}
