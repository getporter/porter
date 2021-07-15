package porter

import (
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/linter"
	"get.porter.sh/porter/pkg/mixin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_Lint(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")

	mixins := p.Mixins.(*mixin.TestMixinProvider)
	mixins.LintResults = linter.Results{
		{
			Level: linter.LevelError,
		},
	}

	var opts LintOptions
	err := opts.Validate(p.Context)
	require.NoError(t, err, "Validate failed")

	results, err := p.Lint(opts)
	require.NoError(t, err, "Lint failed")
	assert.Len(t, results, 1, "Lint returned the wrong number of results")
}

func TestPorter_PrintLintResults(t *testing.T) {
	lintResults := linter.Results{
		{
			Level: linter.LevelError,
			Location: linter.Location{
				Action:          "install",
				Mixin:           "exec",
				StepNumber:      2,
				StepDescription: "Install Hello World",
			},
			Code:  "exec-100",
			Title: "bash -c argument missing wrapping quotes",
			Message: `The bash -c flag argument must be wrapped in quotes, for example
exec:
  description: Say Hello
  command: bash
  flags:
    c: '"echo Hello World"'
`,
			URL: "https://porter.sh/best-practices/exec-mixin/#quoting-escaping-bash-and-yaml",
		},
	}

	testcases := []struct {
		format         string
		wantOutputFile string
	}{
		{"plaintext", "testdata/lint/results.txt"},
		{"json", "testdata/lint/results.json"},
	}
	for _, tc := range testcases {
		t.Run(tc.format, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Teardown()

			p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")

			mixins := p.Mixins.(*mixin.TestMixinProvider)
			mixins.LintResults = lintResults

			var opts LintOptions
			opts.RawFormat = tc.format
			err := opts.Validate(p.Context)
			require.NoError(t, err, "Validate failed")

			err = p.PrintLintResults(opts)
			require.NoError(t, err, "PrintLintResults failed")

			wantOutputB, err := ioutil.ReadFile(tc.wantOutputFile)
			gotOutput := p.TestConfig.TestContext.GetOutput()
			assert.Equal(t, string(wantOutputB), gotOutput, "unexpected output printed")
		})
	}
}
