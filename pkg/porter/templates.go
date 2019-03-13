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

// GetManifestTemplate returns a porter.yaml template file for use in new bundles.
func (t *Templates) GetManifestTemplate() ([]byte, error) {
	return t.box.Find(config.Name)
}

// GetRunScriptTemplate returns a run.sh template for use in new bundles.
func (t *Templates) GetRunScriptTemplate() ([]byte, error) {
	return t.box.Find(path.Base(config.RunScript))
}

// GetSchemaTemplate returns the template manifest schema for the porter manifest.
// Note that is is incomplete and does not include the mixins' schemas.ß
func (t *Templates) GetSchemaTemplate() ([]byte, error) {
	return t.box.Find("schema.json")
}
