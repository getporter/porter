package exec

import (
	"bytes"
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/linter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMixin_LintError(t *testing.T) {
	m := NewTestMixin(t)

	input, err := ioutil.ReadFile("testdata/lint-input.yaml")
	require.NoError(t, err, "could not read lint testdata")
	m.In = bytes.NewReader(input)

	results, err := m.Lint()
	require.NoError(t, err, "Lint failed")
	assert.Len(t, results, 2, "Unexpected number of lint results generated")

	// We expect a warning and an error
	gotInstallWarning := results[0]
	wantInstallWarning := linter.Result{
		Level:   linter.LevelWarning,
		Key:     "c: echo Hello World",
		Code:    CodeEmbeddedBash,
		Title:   "Best Practice: Avoid Embedded Bash",
		Message: "",
		URL:     "https://porter.sh/best-practices/exec-mixin/#use-scripts",
	}
	assert.Equal(t, wantInstallWarning, gotInstallWarning)

	gotInstallError := results[1]
	wantInstallError := linter.Result{
		Level: linter.LevelError,
		Key:   "echo Hello World",
		Code:  CodeBashCArgMissingQuotes,
		Title: "bash -c argument missing wrapping quotes",
		Message: `The bash -c flag argument must be wrapped in quotes, for example
exec:
  description: Say Hello
  command: bash
  flags:
    c: '"echo Hello World"'
`,
		URL: "https://porter.sh/best-practices/exec-mixin/#quoting-escaping-bash-and-yaml",
	}
	assert.Equal(t, wantInstallError, gotInstallError)
}
