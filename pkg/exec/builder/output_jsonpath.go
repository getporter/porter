package builder

import (
	"encoding/json"
	"fmt"

	"github.com/PaesslerAG/jsonpath"
	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

type OutputJsonPath interface {
	Output
	GetJsonPath() string
}

// ProcessJsonPathOutputs evaluates the specified output buffer as JSON, looks through the outputs for
// any that implement the OutputJsonPath and extracts their output.
func ProcessJsonPathOutputs(cxt *context.Context, step StepWithOutputs, stdout string) error {
	outputs := step.GetOutputs()

	if len(outputs) == 0 {
		return nil
	}

	var outputJson interface{}

	for _, o := range outputs {
		output, ok := o.(OutputJsonPath)
		if !ok {
			continue
		}

		outputName := output.GetName()
		outputPath := output.GetJsonPath()
		if outputPath == "" {
			continue
		}

		if cxt.Debug {
			fmt.Fprintf(cxt.Err, "Processing jsonpath output %s...", outputName)
		}

		var valueB []byte

		if outputJson == nil {
			if stdout != "" {
				err := json.Unmarshal([]byte(stdout), &outputJson)
				if err != nil {
					return errors.Wrapf(err, "error unmarshaling stdout as json %s", stdout)
				}
			}
		}

		// Always write an output, even when there isn't json output to parse (like when stdout is empty)
		if outputJson != nil {
			value, err := jsonpath.Get(outputPath, outputJson)
			if err != nil {
				return errors.Wrapf(err, "error evaluating jsonpath %q for output %q against %s", outputPath, outputName, stdout)
			}

			// Only marshal complex types to json, leave strings, numbers and booleans alone
			switch t := value.(type) {
			case map[string]interface{}, []interface{}:
				valueB, err = json.Marshal(value)
				if err != nil {
					return errors.Wrapf(err, "error marshaling jsonpath result %v for output %q", valueB, outputName)
				}
			default:
				valueB = []byte(fmt.Sprintf("%v", t))
			}
		}

		err := cxt.WriteMixinOutputToFile(outputName, valueB)
		if err != nil {
			return errors.Wrapf(err, "error writing mixin output for %q", outputName)
		}
	}

	return nil
}
