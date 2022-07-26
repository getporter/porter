package exec

import (
	"context"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/exec/builder"
	"get.porter.sh/porter/pkg/linter"
	"get.porter.sh/porter/pkg/yaml"
)

// BuildInput represents stdin sent by porter to the build and lint commands
type BuildInput struct {
	// exec mixin doesn't have any buildtime config, so we don't have that field

	// Actions is all the exec actions defined in the manifest
	Actions Actions `yaml:"actions"`
}

const (
	// CodeEmbeddedBash is the linter code for when a bash -c command is found.
	CodeEmbeddedBash linter.Code = "exec-100"

	// CodeBashCArgMissingQuotes is the linter code for when a bash -c flag argument is missing the required wrapping quotes.
	CodeBashCArgMissingQuotes linter.Code = "exec-101"
)

func (m *Mixin) Lint(ctx context.Context) (linter.Results, error) {
	var input BuildInput

	err := builder.LoadAction(ctx, m.Config, "", func(contents []byte) (interface{}, error) {
		err := yaml.Unmarshal(contents, &input)
		return &input, err
	})
	if err != nil {
		return nil, err
	}

	// Right now the only exec invocation we are looking for is
	// 		bash -c "some command"
	// We are looking for:
	//  * using that command at all (WARN)
	//  * missing wrapping quotes around the command (ERROR)
	results := make(linter.Results, 0)

	for _, action := range input.Actions {
		for stepNumber, step := range action.Steps {
			if step.Command != "bash" {
				continue
			}

			var embeddedBashFlag *builder.Flag
			for _, flag := range step.Flags {
				if flag.Name == "c" {
					embeddedBashFlag = &flag
					break
				}
			}

			if embeddedBashFlag == nil {
				continue
			}

			// Found embedded bash 🚨
			// Check for wrapping quotes, if missing -> hard error, otherwise just warn
			result := linter.Result{
				Level: linter.LevelWarning,
				Code:  CodeEmbeddedBash,
				Location: linter.Location{
					Action:          action.Name,
					Mixin:           "exec",
					StepNumber:      stepNumber + 1, // We index from 1 for natural counting, 1st, 2nd, etc.
					StepDescription: step.Description,
				},
				Title:   "Best Practice: Avoid Embedded Bash",
				Message: "",
				URL:     "https://getporter.org/best-practices/exec-mixin/#use-scripts",
			}
			results = append(results, result)

			for _, bashCmd := range embeddedBashFlag.Values {
				if (!strings.HasPrefix(bashCmd, `"`) || !strings.HasSuffix(bashCmd, `"`)) &&
					(!strings.HasPrefix(bashCmd, `'`) || !strings.HasSuffix(bashCmd, `'`)) {
					result := linter.Result{
						Level: linter.LevelError,
						Code:  CodeBashCArgMissingQuotes,
						Location: linter.Location{
							Action:          action.Name,
							Mixin:           "exec",
							StepNumber:      stepNumber + 1,
							StepDescription: step.Description,
						},
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
					results = append(results, result)
					break
				}
			}
		}
	}

	return results, nil
}

func (m *Mixin) PrintLintResults(ctx context.Context) error {
	results, err := m.Lint(ctx)
	if err != nil {
		return err
	}

	b, err := encoding.MarshalJson(results)
	if err != nil {
		return fmt.Errorf("could not marshal lint results %#v: %w", results, err)
	}

	// Print the results as json to stdout for Porter to read
	resultsJson := string(b)
	fmt.Fprintln(m.Config.Out, resultsJson)

	return nil
}
