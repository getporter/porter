package parameters

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/credentials"
	"gopkg.in/yaml.v2"
)

// ParseVariableAssignments converts a string array of variable assignments
// into a map of keys and values
// Example:
// [a=b c=abc1232=== d=banana d=pineapple] becomes map[a:b c:abc1232=== d:[pineapple]]
func ParseVariableAssignments(params []string) (map[string]string, error) {
	variables := make(map[string]string)
	for _, p := range params {

		parts := strings.SplitN(p, "=", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid parameter (%s), must be in name=value format", p)
		}

		variable := strings.TrimSpace(parts[0])
		if variable == "" {
			return nil, fmt.Errorf("invalid parameter (%s), variable name is required", p)
		}
		value := strings.TrimSpace(parts[1])

		variables[variable] = value
	}

	return variables, nil
}

// Load a ParameterSet from a file at a given path.
//
// It does not load the individual parameters.
func Load(path string) (*ParameterSet, error) {
	pset := &ParameterSet{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return pset, err
	}
	return pset, yaml.Unmarshal(data, pset)
}

// Validate compares the given parameters with the spec.
//
// This will result in an error only when the following conditions are true:
// - a parameter in the spec is not present in the given set
// - the parameter is required
//
// It is allowed for spec to specify both an env var and a file. In such case, if
// the given set provides either, it will be considered valid.
func Validate(given credentials.Set, spec map[string]bundle.Parameter) error {
	for name, param := range spec {
		if !isValidParam(given, name) && param.Required {
			return fmt.Errorf("bundle requires parameter for %s", name)
		}
	}
	return nil
}

func isValidParam(haystack credentials.Set, needle string) bool {
	for name := range haystack {
		if name == needle {
			return true
		}
	}
	return false
}
