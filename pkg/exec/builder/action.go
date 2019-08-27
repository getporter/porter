package builder

import (
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
