package builder

import (
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
)

type OutputFile interface {
	Output
	GetFilePath() string
}

// ProcessFileOutputs makes the contents of a file specified by any OutputFile interface available as an output.
func ProcessFileOutputs(cxt *portercontext.Context, step StepWithOutputs) error {
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

		if cxt.Debug {
			fmt.Fprintf(cxt.Err, "Processing file output %s...", outputName)
		}

		valueB, err := cxt.FileSystem.ReadFile(outputPath)
		if err != nil {
			return fmt.Errorf("error evaluating filepath %q for output %q: %w", outputPath, outputName, err)
		}

		err = cxt.WriteMixinOutputToFile(outputName, valueB)
		if err != nil {
			return fmt.Errorf("error writing mixin output for %q: %w", outputName, err)
		}
	}

	return nil
}
