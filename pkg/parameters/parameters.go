package parameters

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/schema"
)

const (
	// SchemaVersion represents the version associated with the schema
	// for parameter set documents.
	SchemaVersion = schema.Version("1.0.1")

	//
	DEFAULT_STRATEGY = "default-strategy"
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

// DefaultStrategy is the strategy used to store non-sensitive parameters
func DefaultStrategy(name string, value string) secrets.Strategy {
	return secrets.Strategy{
		Name:   name,
		Source: secrets.Source{Key: DEFAULT_STRATEGY, Value: value},
		Value:  value,
	}
}
