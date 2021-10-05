package runtime

import (
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/yaml"
)

type ActionInput struct {
	action string
	Steps  []*manifest.Step `yaml:"steps"`
}

// MarshalYAML marshals the step nested under the action
// install:
// - helm3:
//   ...
// Solution from https://stackoverflow.com/a/42547226
func (a *ActionInput) MarshalYAML() (interface{}, error) {
	// encode the original
	b, err := yaml.Marshal(a.Steps)
	if err != nil {
		return nil, err
	}

	// decode it back to get a map
	var tmp interface{}
	err = yaml.Unmarshal(b, &tmp)
	if err != nil {
		return nil, err
	}
	stepMap := tmp.([]interface{})
	actionMap := map[string]interface{}{
		string(a.action): stepMap,
	}
	return actionMap, nil
}
