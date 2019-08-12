package runtime

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

// GetDependencyOutputsDir determines the outputs directory for a dependency
func GetDependencyOutputsDir(alias string) string {
	return filepath.Join(BundleDependenciesDir, alias, "outputs")
}

// ReadDependencyOutputValue reads the dependency's output using the alias for the dependency from the
// specified output file (name).
func ReadDependencyOutputValue(c *context.Context, alias string, name string) (string, error) {
	outputFile := filepath.Join(GetDependencyOutputsDir(alias), name)

	exists, err := c.FileSystem.Exists(outputFile)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read %s", outputFile)
	}
	if !exists {
		return "", errors.Errorf("outputs file %s does not exist", outputFile)
	}

	b, err := c.FileSystem.ReadFile(outputFile)
	if err != nil {
		return "", errors.Errorf("unable to read %s", outputFile)
	}

	var output map[string]interface{}
	err = json.Unmarshal(b, &output)
	if err != nil {
		return "", errors.Wrapf(err, "error unmarshaling output %s.%s from data %q", alias, name, string(b))
	}

	switch value := output["value"].(type) {
	case string:
		return value, nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}
