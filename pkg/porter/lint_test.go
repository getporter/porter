package porter

import (
	"context"
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/linter"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_Lint_ChecksManifestSchemaVersion(t *testing.T) {
	testcases := []struct {
		name          string
		schemaVersion string
		wantErr       string
	}{
		{name: "valid version", schemaVersion: manifest.DefaultSchemaVersion.String()},
		{name: "invalid version", schemaVersion: "", wantErr: "invalid schema version"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			// Make a bundle with the specified schemaVersion
			p.TestConfig.TestContext.AddTestFileFromRoot("tests/testdata/mybuns/porter.yaml", "porter.yaml")
			e := yaml.NewEditor(p.Context)
			require.NoError(t, e.ReadFile("porter.yaml"))
			require.NoError(t, e.SetValue("schemaVersion", tc.schemaVersion))
			require.NoError(t, e.WriteFile("porter.yaml"))

			_, err := p.Lint(context.Background(), LintOptions{File: "porter.yaml"})
			if tc.wantErr == "" {
				require.NoError(t, err, "Lint failed")
			} else {
				tests.RequireErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

func TestPorter_Lint(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

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

	results, err := p.Lint(context.Background(), opts)
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
			URL: "https://getporter.org/best-practices/exec-mixin/#quoting-escaping-bash-and-yaml",
		},
	}

	testcases := []struct {
		format         string
		wantOutputFile string
		linterResults  linter.Results
	}{
		{"plaintext", "testdata/lint/results.txt", lintResults},
		{"json", "testdata/lint/results.json", lintResults},
		{"plaintext", "testdata/lint/success.txt", linter.Results{}},
	}
	for _, tc := range testcases {
		t.Run(tc.format, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")

			mixins := p.Mixins.(*mixin.TestMixinProvider)
			mixins.LintResults = tc.linterResults

			var opts LintOptions
			opts.RawFormat = tc.format
			err := opts.Validate(p.Context)
			require.NoError(t, err, "Validate failed")

			err = p.PrintLintResults(context.Background(), opts)
			require.NoError(t, err, "PrintLintResults failed")

			wantOutputB, err := ioutil.ReadFile(tc.wantOutputFile)
			require.NoError(t, err, "Reading output file failed")
			gotOutput := p.TestConfig.TestContext.GetOutput()
			assert.Equal(t, string(wantOutputB), gotOutput, "unexpected output printed")
		})
	}
}
