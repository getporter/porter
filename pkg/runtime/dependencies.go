package runtime

import (
	"path/filepath"

	"get.porter.sh/porter/pkg/portercontext"
	"github.com/pkg/errors"
)

const (
	BundleDependenciesDir = "/cnab/app/dependencies"
)

func GetDependencyDefinitionPath(alias string) string {
	return filepath.Join(BundleDependenciesDir, alias, "bundle.json")
}

func GetDependencyDefinition(c *portercontext.Context, alias string) ([]byte, error) {
	f := GetDependencyDefinitionPath(alias)
	data, err := c.FileSystem.ReadFile(f)
	return data, errors.Wrapf(err, "error reading bundle definition for %s at %s", alias, f)
}
