//go:generate packr2

package templates

import (
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
	return packr.New("get.porter.sh/porter/pkg/templates", "./templates")
}

// GetManifest returns a porter.yaml template file for use in new bundles.
func (t *Templates) GetManifest() ([]byte, error) {
	return t.box.Find("create/porter.yaml")
}

// GetHelpers returns a helpers.sh template file for use in new bundles.
func (t *Templates) GetManifestHelpers() ([]byte, error) {
	return t.box.Find("create/helpers.sh")
}

// GetReadme returns a README.md file for use in new bundles.
func (t *Templates) GetReadme() ([]byte, error) {
	return t.box.Find("create/README.md")
}

// GetGitignore returns a .gitignore file for use in new bundles.
func (t *Templates) GetGitignore() ([]byte, error) {
	return t.box.Find("create/.gitignore")
}

// GetDockerignore returns a .dockerignore file for use in new bundles.
func (t *Templates) GetDockerignore() ([]byte, error) {
	return t.box.Find("create/.dockerignore")
}

// GetDockerfileTemplate returns a Dockerfile.tmpl file for use in new bundles.
func (t *Templates) GetDockerfileTemplate() ([]byte, error) {
	return t.box.Find("create/Dockerfile.tmpl")
}

// GetRunScript returns a run script template for invocation images.
func (t *Templates) GetRunScript() ([]byte, error) {
	return t.box.Find("build/cnab/app/run")
}

// GetSchema returns the template manifest schema for the porter manifest.
// Note that is is incomplete and does not include the mixins' schemas.ß
func (t *Templates) GetSchema() ([]byte, error) {
	return t.box.Find("schema.json")
}

// GetDockerfile returns the default Dockerfile for invocation images.
func (t *Templates) GetDockerfile() ([]byte, error) {
	return t.box.Find("build/Dockerfile")
}
