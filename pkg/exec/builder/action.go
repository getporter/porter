package builder

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"

	"get.porter.sh/porter/pkg/runtime"
	"get.porter.sh/porter/pkg/tracing"
	"get.porter.sh/porter/pkg/yaml"
)

// UnmarshalAction handles unmarshaling any action, given a pointer to a slice of Steps.
// Iterate over the results and cast back to the Steps to use the results.
//  var steps []Step
//	results, err := UnmarshalAction(unmarshal, &steps)
//	if err != nil {
//		return err
//	}
//
//	for _, result := range results {
//		step := result.(*[]Step)
//		a.Steps = append(a.Steps, *step...)
//	}
func UnmarshalAction(unmarshal func(interface{}) error, builder BuildableAction) (map[string][]interface{}, error) {
	actionMap := map[string][]interface{}{}
	err := unmarshal(&actionMap)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal yaml into an action map of exec steps: %w", err)
	}

	return unmarshalActionMap(actionMap, builder)
}

func unmarshalActionMap(actionMap map[string][]interface{}, builder BuildableAction) (map[string][]interface{}, error) {
	results := make(map[string][]interface{})
	for actionIndex, stepMaps := range actionMap {
		// Figure out the string representation of the action
		// examples:
		//   install: -> "install"
		//   true: -> "true" YAML is weird, this is why we use Sprintf and not .(string)
		name := fmt.Sprintf("%v", actionIndex)

		// Unmarshal the steps
		b, err := yaml.Marshal(stepMaps)
		if err != nil {
			return nil, err
		}

		steps := builder.MakeSteps()
		err = yaml.Unmarshal(b, steps)
		if err != nil {
			return nil, err
		}

		result, ok := results[name]
		if !ok {
			result = make([]interface{}, 0, 1)
		}
		results[name] = append(result, steps)
	}

	return results, nil
}

// LoadAction reads input from stdin or a command file and uses the specified unmarshal function
// to unmarshal it into a typed Action.
// The unmarshal function is responsible for calling yaml.Unmarshal and passing in a reference to an appropriate
// Action instance.
//
// Example:
//   var action Action
//	 err := builder.LoadAction(m.Context, opts.File, func(contents []byte) (interface{}, error) {
//		 err := yaml.Unmarshal(contents, &action)
//		 return &action, err
//	 })
func LoadAction(ctx context.Context, cfg runtime.RuntimeConfig, commandFile string, unmarshal func([]byte) (interface{}, error)) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	contents, err := readInputFromStdinOrFile(cfg, commandFile)
	if err != nil {
		return span.Error(err)
	}

	_, err = unmarshal(contents)
	if err != nil {
		return span.Error(fmt.Errorf("could not unmarshal input:\n %s: %w", string(contents), err))
	}

	return nil
}

func readInputFromStdinOrFile(cfg runtime.RuntimeConfig, commandFile string) ([]byte, error) {
	var b []byte
	var err error
	if commandFile == "" {
		reader := bufio.NewReader(cfg.In)
		b, err = ioutil.ReadAll(reader)
	} else {
		b, err = cfg.FileSystem.ReadFile(commandFile)
	}

	if err != nil {
		source := "STDIN"
		if commandFile == "" {
			source = commandFile
		}
		return nil, fmt.Errorf("could not load input from %s: %w", source, err)
	}
	return b, nil
}
