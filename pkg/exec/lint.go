package exec

import (
	"encoding/json"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/exec/builder"
	"get.porter.sh/porter/pkg/linter"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
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

func (m *Mixin) Lint() (linter.Results, error) {
	var input BuildInput

	err := builder.LoadAction(m.Context, "", func(contents []byte) (interface{}, error) {
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
		for _, step := range action.Steps {
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

			// Derive key to be used to locate coordinates in the manifest.
			// The embedded bash flag, 'c' is not unique enough,
			// so set to the unique step description, or, if values non-empty,
			// the first value
			key := step.Description
			if len(embeddedBashFlag.Values) > 0 {
				key = embeddedBashFlag.Values[0]
			}
			// Found embedded bash 🚨
			// Check for wrapping quotes, if missing -> hard error, otherwise just warn
			result := linter.Result{
				Level:   linter.LevelWarning,
				Code:    CodeEmbeddedBash,
				Key:     key,
				Title:   "Best Practice: Avoid Embedded Bash",
				Message: "",
				URL:     "https://porter.sh/best-practices/exec-mixin/#use-scripts",
			}
			results = append(results, result)

			for _, bashCmd := range embeddedBashFlag.Values {
				if (!strings.HasPrefix(bashCmd, `"`) || !strings.HasSuffix(bashCmd, `"`)) &&
					(!strings.HasPrefix(bashCmd, `'`) || !strings.HasSuffix(bashCmd, `'`)) {
					result := linter.Result{
						Level: linter.LevelError,
						Code:  CodeBashCArgMissingQuotes,
						Key:   bashCmd,
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
					results = append(results, result)
					break
				}
			}
		}
	}

	return results, nil
}

func (m *Mixin) PrintLintResults() error {
	results, err := m.Lint()
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "could not marshal lint results %#v", results)
	}

	// Print the results as json to stdout for Porter to read
	resultsJson := string(b)
	fmt.Fprintln(m.Out, resultsJson)

	return nil
}
