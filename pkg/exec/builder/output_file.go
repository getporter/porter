package builder

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/runtime"
	"get.porter.sh/porter/pkg/tracing"
)

type OutputFile interface {
	Output
	GetFilePath() string
}

// ProcessFileOutputs makes the contents of a file specified by any OutputFile interface available as an output.
func ProcessFileOutputs(ctx context.Context, cfg runtime.RuntimeConfig, step StepWithOutputs) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	outputs := step.GetOutputs()

	if len(outputs) == 0 {
		return nil
	}

	for _, o := range outputs {
		output, ok := o.(OutputFile)
		if !ok {
			continue
		}

		outputName := output.GetName()
		outputPath := output.GetFilePath()
		if outputPath == "" {
			continue
		}

		span.Debugf("Processing file output %s...", outputName)

		valueB, err := cfg.FileSystem.ReadFile(outputPath)
		if err != nil {
			return fmt.Errorf("error evaluating filepath %q for output %q: %w", outputPath, outputName, err)
		}

		err = cfg.WriteMixinOutputToFile(outputName, valueB)
		if err != nil {
			return fmt.Errorf("error writing mixin output for %q: %w", outputName, err)
		}
	}

	return nil
}
