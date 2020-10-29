package builder

import (
	"fmt"
	"regexp"
	"strings"

	"get.porter.sh/porter/pkg/context"
	"github.com/pkg/errors"
)

type OutputRegex interface {
	Output
	GetRegex() string
}

// ProcessJsonPathOutputs looks through the outputs for any that implement the OutputRegex,
// applies the regular expression to the output buffer and extracts their output.
func ProcessRegexOutputs(cxt *context.Context, step StepWithOutputs, stdout string) error {
	outputs := step.GetOutputs()

	if len(outputs) == 0 {
		return nil
	}

	for _, o := range outputs {
		output, ok := o.(OutputRegex)
		if !ok {
			continue
		}

		outputName := output.GetName()
		outputRegex := output.GetRegex()
		if outputRegex == "" {
			continue
		}

		if cxt.Debug {
			fmt.Fprintf(cxt.Err, "Processing regex output %s...", outputName)
		}

		r, err := regexp.Compile(outputRegex)
		if err != nil {
			return errors.Wrapf(err, "invalid regular expression %q for output %q", outputRegex, outputName)
		}

		// Find every submatch / capture and put it on its own line in the output file
		results := r.FindAllStringSubmatch(stdout, -1)
		var matches []string
		for _, result := range results {
			if len(result) > 1 { // Skip the first element which is the full match, we only want the capture groups
				matches = append(matches, result[1:]...)
			}
		}
		value := strings.Join(matches, "\n")
		err = cxt.WriteMixinOutputToFile(outputName, []byte(value))
		if err != nil {
			return errors.Wrapf(err, "error writing mixin output for %q", outputName)
		}
	}

	return nil
}
