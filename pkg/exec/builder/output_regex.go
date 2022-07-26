package builder

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"get.porter.sh/porter/pkg/runtime"
	"get.porter.sh/porter/pkg/tracing"
)

type OutputRegex interface {
	Output
	GetRegex() string
}

// ProcessRegexOutputs looks through the outputs for any that implement the OutputRegex,
// applies the regular expression to the output buffer and extracts their output.
func ProcessRegexOutputs(ctx context.Context, cfg runtime.RuntimeConfig, step StepWithOutputs, stdout string) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

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

		span.Debugf("Processing regex output %s...", outputName)

		r, err := regexp.Compile(outputRegex)
		if err != nil {
			return span.Error(fmt.Errorf("invalid regular expression %q for output %q: %w", outputRegex, outputName, err))
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
		err = cfg.WriteMixinOutputToFile(outputName, []byte(value))
		if err != nil {
			return span.Error(fmt.Errorf("error writing mixin output for %q: %w", outputName, err))
		}
	}

	return nil
}
