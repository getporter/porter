package builder

import (
	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

type OutputFile interface {
	Output
	GetFilePath() string
}

// ProcessFileOutputs makes the contents of a file specified by any OutputFile interface available as an output.
func ProcessFileOutputs(cxt *context.Context, step StepWithOutputs) error {
	outputs := step.GetOutputs()

	if len(outputs) == 0 {
		return nil
	}

	for _, o := range outputs {
		output, ok := o.(OutputFile)
		if !ok {
			continue
		}

		outputPath := output.GetFilePath()
		outputName := output.GetName()

		valueB, err := cxt.FileSystem.ReadFile(outputPath)
		if err != nil {
			return errors.Wrapf(err, "error evaluating filepath %q for output %q", outputPath, outputName)
		}

		err = cxt.WriteMixinOutputToFile(outputName, valueB)
		if err != nil {
			return errors.Wrapf(err, "error writing mixin output for %q", outputName)
		}
	}

	return nil
}
