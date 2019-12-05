package builder

import (
	"bufio"
	"fmt"
	"io/ioutil"

	"get.porter.sh/porter/pkg/context"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
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
func UnmarshalAction(unmarshal func(interface{}) error, steps interface{}) ([]interface{}, error) {
	actionMap := map[interface{}][]interface{}{}
	err := unmarshal(&actionMap)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal yaml into an action map of exec steps")
	}

	var result []interface{}
	for _, stepMaps := range actionMap {
		b, err := yaml.Marshal(stepMaps)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(b, steps)
		if err != nil {
			return nil, err
		}
		result = append(result, steps)
	}

	return result, nil
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
func LoadAction(cxt *context.Context, commandFile string, unmarshal func([]byte) (interface{}, error)) error {
	contents, err := readInputFromStdinOrFile(cxt, commandFile)
	if err != nil {
		return err
	}

	result, err := unmarshal(contents)
	if cxt.Debug {
		fmt.Fprintf(cxt.Err, "DEBUG Parsed Input:\n%#v\n", result)
	}
	return errors.Wrapf(err, "could unmarshal input:\n %s", string(contents))
}

func readInputFromStdinOrFile(cxt *context.Context, commandFile string) ([]byte, error) {
	var b []byte
	var err error
	if commandFile == "" {
		reader := bufio.NewReader(cxt.In)
		b, err = ioutil.ReadAll(reader)
	} else {
		b, err = cxt.FileSystem.ReadFile(commandFile)
	}

	if err != nil {
		source := "STDIN"
		if commandFile == "" {
			source = commandFile
		}
		return nil, errors.Wrapf(err, "could not load input from %s", source)
	}
	return b, nil
}
