package porter

import (
	"context"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/linter"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle"
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
			e := yaml.NewEditor(p.FileSystem)
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
			URL: "https://porter.sh/best-practices/exec-mixin/#quoting-escaping-bash-and-yaml",
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

			wantOutputB, err := os.ReadFile(tc.wantOutputFile)
			require.NoError(t, err, "Reading output file failed")
			gotOutput := p.TestConfig.TestContext.GetOutput()
			assert.Equal(t, string(wantOutputB), gotOutput, "unexpected output printed")
		})
	}
}

func TestPorter_Lint_DependencyMappings(t *testing.T) {
	const manifestYaml = `schemaVersion: 1.0.0
name: has-deps
version: 0.1.0
registry: "localhost:5000"

dependencies:
  requires:
    - name: mysql
      bundle:
        reference: localhost:5000/mysql:v1.0.0
      parameters:
        NOT_A_PARAM: "value"
      credentials:
        NOT_A_CRED: "value"
    - name: unresolvable
      bundle:
        reference: localhost:5000/unresolvable:v1.0.0

mixins:
  - exec

install:
  - exec:
      description: "Install"
      command: echo
      arguments:
        - "hello"

uninstall:
  - exec:
      description: "Uninstall"
      command: echo
      arguments:
        - "goodbye"
`

	p := NewTestPorter(t)
	defer p.Close()

	require.NoError(t, p.TestConfig.TestContext.AddTestFileContents([]byte(manifestYaml), "porter.yaml"))

	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		"localhost:5000/mysql:v1.0.0": {Bundle: bundle.Bundle{
			Name:    "mysql",
			Version: "1.0.0",
			Parameters: map[string]bundle.Parameter{
				"REAL_PARAM": {Definition: "real-param"},
			},
			Credentials: map[string]bundle.Credential{
				"REAL_CRED": {},
			},
		}},
	})

	results, err := p.Lint(context.Background(), LintOptions{File: "porter.yaml"})
	require.NoError(t, err, "Lint failed")

	codes := make([]linter.Code, 0, len(results))
	for _, r := range results {
		codes = append(codes, r.Code)
	}
	assert.ElementsMatch(t, []linter.Code{"porter-103", "porter-104", "porter-105"}, codes, "unexpected lint results: %v", results)

	for _, r := range results {
		switch r.Code {
		case "porter-103":
			assert.Equal(t, linter.LevelError, r.Level)
			assert.Contains(t, r.Message, "dependencies.mysql.parameters.NOT_A_PARAM")
		case "porter-104":
			assert.Equal(t, linter.LevelError, r.Level)
			assert.Contains(t, r.Message, "dependencies.mysql.credentials.NOT_A_CRED")
		case "porter-105":
			assert.Equal(t, linter.LevelWarning, r.Level)
			assert.Contains(t, r.Message, "unresolvable")
		}
	}
}

func TestPorter_PrintLintResults_Warning(t *testing.T) {
	lintResults := linter.Results{
		{
			Level: linter.LevelWarning,
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
		linterResults  linter.Results
	}{
		{"plaintext", "testdata/lint/results_warning.txt", lintResults},
		{"json", "testdata/lint/results_warning.json", lintResults},
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

			wantOutputB, err := os.ReadFile(tc.wantOutputFile)
			require.NoError(t, err, "Reading output file failed")
			gotOutput := p.TestConfig.TestContext.GetOutput()
			assert.Equal(t, string(wantOutputB), gotOutput, "unexpected output printed")
		})
	}
}
