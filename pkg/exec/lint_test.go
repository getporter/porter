package exec

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/linter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMixin_Lint(t *testing.T) {
	ctx := context.Background()
	m := NewTestMixin(t)

	input, err := ioutil.ReadFile("testdata/lint-input.yaml")
	require.NoError(t, err, "could not read lint testdata")
	m.Config.In = bytes.NewReader(input)

	results, err := m.Lint(ctx)
	require.NoError(t, err, "Lint failed")
	assert.Len(t, results, 2, "Unexpected number of lint results generated")

	var gotInstallError linter.Result
	for _, r := range results {
		if r.Location.Action == "install" && r.Code == CodeBashCArgMissingQuotes {
			gotInstallError = r
		}
	}
	wantInstallError := linter.Result{
		Level: linter.LevelError,
		Location: linter.Location{
			Action:          "install",
			Mixin:           "exec",
			StepNumber:      2,
			StepDescription: "Install Hello World",
		},
		Code:  CodeBashCArgMissingQuotes,
		Title: "bash -c argument missing wrapping quotes",
		Message: `The bash -c flag argument must be wrapped in quotes, for example
exec:
  description: Say Hello
  command: bash
  flags:
    c: '"echo Hello World"'
`,
		URL: "https://getporter.org/best-practices/exec-mixin/#quoting-escaping-bash-and-yaml",
	}
	assert.Equal(t, wantInstallError, gotInstallError)
}
