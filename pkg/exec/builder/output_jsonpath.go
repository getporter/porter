package builder

import (
	"bytes"
	"encoding/json"
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
	"github.com/PaesslerAG/jsonpath"
)

type OutputJsonPath interface {
	Output
	GetJsonPath() string
}

// ProcessJsonPathOutputs evaluates the specified output buffer as JSON, looks through the outputs for
// any that implement the OutputJsonPath and extracts their output.
func ProcessJsonPathOutputs(cxt *portercontext.Context, step StepWithOutputs, stdout string) error {
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
			fmt.Fprintf(cxt.Err, "Processing jsonpath output %s using query %s against document\n%s\n", outputName, outputPath, stdout)
		}

		var valueB []byte

		if outputJson == nil {
			if stdout != "" {
				d := json.NewDecoder(bytes.NewBuffer([]byte(stdout)))
				d.UseNumber()
				err := d.Decode(&outputJson)
				if err != nil {
					return fmt.Errorf("error unmarshaling stdout as json %s: %w", stdout, err)
				}
			}
		}

		// Always write an output, even when there isn't json output to parse (like when stdout is empty)
		if outputJson != nil {
			value, err := jsonpath.Get(outputPath, outputJson)
			if err != nil {
				return fmt.Errorf("error evaluating jsonpath %q for output %q against %s: %w", outputPath, outputName, stdout, err)
			}

			// Only marshal complex types to json, leave strings, numbers and booleans alone
			switch t := value.(type) {
			case map[string]interface{}, []interface{}:
				valueB, err = json.Marshal(value)
				if err != nil {
					return fmt.Errorf("error marshaling jsonpath result %v for output %q: %w", valueB, outputName, err)
				}
			default:
				valueB = []byte(fmt.Sprintf("%v", t))
			}
		}

		err := cxt.WriteMixinOutputToFile(outputName, valueB)
		if err != nil {
			return fmt.Errorf("error writing mixin output for %q: %w", outputName, err)
		}
	}

	return nil
}
