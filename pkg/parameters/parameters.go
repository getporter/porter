package parameters

import (
	"fmt"
	"log"
	"strings"

	"github.com/cnabio/cnab-go/schema"
)

const (
	// SchemaVersion represents the version associated with the schema
	// for parameter set documents.
	SchemaVersion = schema.Version("1.0.0")

	// PorterInternal is a string that can be used to designate a parameter
	// as internal to Porter

)

// ParseVariableAssignments converts a string array of variable assignments
// into a map of keys and values
// Example:
// [a=b c=abc1232=== d=banana d=pineapple] becomes map[a:b c:abc1232=== d:[pineapple]]
func ParseVariableAssignments(params []string) (map[string]string, error) {
	variables := make(map[string]string)
	for _, p := range params {
		log.Println(p)

		parts := strings.SplitN(p, "=", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid parameter (%s), must be in name=value format", p)
		}
		log.Println(parts)

		variable := strings.TrimSpace(parts[0])
		if variable == "" {
			return nil, fmt.Errorf("invalid parameter (%s), variable name is required", p)
		}
		log.Println(variable)
		value := strings.TrimSpace(parts[1])

		log.Println(value)
		variables[variable] = value
	}

	return variables, nil
}
