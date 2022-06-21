package templates

import (
	"embed"
	"fmt"

	"get.porter.sh/porter/pkg/config"
)

//go:embed templates/*
var fs embed.FS

// Workaround until go:embed can include hidden files
// https://github.com/golang/go/issues/43854
//go:embed templates/create/.dockerignore
var dockerignore []byte

//go:embed templates/create/.gitignore
var gitignore []byte

type Templates struct {
	*config.Config
	fs embed.FS
}

func NewTemplates(c *config.Config) *Templates {
	return &Templates{
		Config: c,
		fs:     fs,
	}
}

// GetManifest returns a porter.yaml template file for use in new bundles.
func (t *Templates) GetManifest() ([]byte, error) {
	return t.fs.ReadFile("templates/create/porter.yaml")
}

// GetHelpers returns a helpers.sh template file for use in new bundles.
func (t *Templates) GetManifestHelpers() ([]byte, error) {
	return t.fs.ReadFile("templates/create/helpers.sh")
}

// GetReadme returns a README.md file for use in new bundles.
func (t *Templates) GetReadme() ([]byte, error) {
	return t.fs.ReadFile("templates/create/README.md")
}

// GetGitignore returns a .gitignore file for use in new bundles.
func (t *Templates) GetGitignore() ([]byte, error) {
	return gitignore, nil
}

// GetDockerignore returns a .dockerignore file for use in new bundles.
func (t *Templates) GetDockerignore() ([]byte, error) {
	return dockerignore, nil
}

// GetDockerfileTemplate returns a template.Dockerfile file for use in new bundles.
func (t *Templates) GetDockerfileTemplate() ([]byte, error) {
	tmpl := fmt.Sprintf("templates/create/template.%s.Dockerfile", t.GetBuildDriver())
	return t.fs.ReadFile(tmpl)
}

// GetRunScript returns a run script template for invocation images.
func (t *Templates) GetRunScript() ([]byte, error) {
	return t.fs.ReadFile("templates/build/cnab/app/run")
}

// GetSchema returns the template manifest schema for the porter manifest.
// Note that it is incomplete and does not include the mixins' schemas.
func (t *Templates) GetSchema() ([]byte, error) {
	return t.fs.ReadFile("templates/schema.json")
}

// GetDockerfile returns the default Dockerfile for invocation images.
func (t *Templates) GetDockerfile() ([]byte, error) {
	tmpl := fmt.Sprintf("templates/build/%s.Dockerfile", t.GetBuildDriver())
	return t.fs.ReadFile(tmpl)
}

// GetCredentialSetJSON returns a credential-set.schema.json template file to define new credential set.
func (t *Templates) GetCredentialSetJSON() ([]byte, error) {
	return t.fs.ReadFile("templates/credentials/create/credential-set.json")
}

// GetCredentialSetYAML returns a credential-set.yaml template file to define new credential set.
func (t *Templates) GetCredentialSetYAML() ([]byte, error) {
	return t.fs.ReadFile("templates/credentials/create/credential-set.yaml")
}

// GetParameterSetJSON returns a parameter-set.schema.json template file to define new parameter set.
func (t *Templates) GetParameterSetJSON() ([]byte, error) {
	return t.fs.ReadFile("templates/parameters/create/parameter-set.json")
}

// GetParameterSetYAML returns a parameter-set.yaml template file to define new parameter set.
func (t *Templates) GetParameterSetYAML() ([]byte, error) {
	return t.fs.ReadFile("templates/parameters/create/parameter-set.yaml")
}
