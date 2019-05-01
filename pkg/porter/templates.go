//go:generate packr2

package porter

import (
	"path"

	"github.com/deislabs/porter/pkg/config"
	"github.com/gobuffalo/packr/v2"
)

type Templates struct {
	box *packr.Box
}

func NewTemplates() *Templates {
	return &Templates{
		box: NewTemplatesBox(),
	}
}

// NewSchemas creates or retrieves the packr box with the porter template files.
func NewTemplatesBox() *packr.Box {
	return packr.New("github.com/deislabs/porter/pkg/porter/templates", "./templates")
}

// GetManifest returns a porter.yaml template file for use in new bundles.
func (t *Templates) GetManifest() ([]byte, error) {
	return t.box.Find(config.Name)
}

// GetRunScript returns a run.sh template for use in new bundles.
func (t *Templates) GetRunScript() ([]byte, error) {
	return t.box.Find(path.Base(config.RunScript))
}

// GetSchema returns the template manifest schema for the porter manifest.
// Note that is is incomplete and does not include the mixins' schemas.ß
func (t *Templates) GetSchema() ([]byte, error) {
	return t.box.Find("schema.json")
}

// GetDockerfile returns the default Dockerfile for invocation images.
func (t *Templates) GetDockerfile() ([]byte, error) {
	return t.box.Find("Dockerfile")
}
