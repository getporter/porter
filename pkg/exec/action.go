package exec

import (
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type Action struct {
	Steps []Step // using UnmarshalYAML so that we don't need a custom type per action
}

// UnmarshalYAML takes any yaml in this form
// ACTION:
// - exec: ...
// and puts the steps into the Action.Steps field
func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var steps []Step
	results, err := UnmarshalAction(unmarshal, &steps)
	if err != nil {
		return err
	}

	for _, result := range results {
		step := result.(*[]Step)
		a.Steps = append(a.Steps, *step...)
	}
	return nil
}

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

type Step struct {
	Instruction `yaml:"exec"`
}

type Instruction struct {
	Description string   `yaml:"description"`
	Command     string   `yaml:"command"`
	Arguments   []string `yaml:"arguments,omitempty"`
	Flags       Flags    `yaml:"flags,omitempty"`
}
