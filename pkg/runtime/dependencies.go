package runtime

import (
	"fmt"
	"path"

	"get.porter.sh/porter/pkg/portercontext"
)

const (
	BundleDependenciesDir = "/cnab/app/dependencies"
)

func GetDependencyDefinitionPath(alias string) string {
	// Must be a unix path
	return path.Join(BundleDependenciesDir, alias, "bundle.json")
}

func GetDependencyDefinition(c *portercontext.Context, alias string) ([]byte, error) {
	f := GetDependencyDefinitionPath(alias)
	data, err := c.FileSystem.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("error reading bundle definition for %s at %s: %w", alias, f, err)
	}
	return data, nil
}
