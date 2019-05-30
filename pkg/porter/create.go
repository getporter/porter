package porter

import (
	"fmt"
	"github.com/deislabs/porter/pkg/config"
)

func (p *Porter) Create() error {
	fmt.Fprintln(p.Out, "creating porter configuration in the current directory")

	err := p.CopyTemplate(p.Templates.GetManifest, config.Name)
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetReadme, "README.md")
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetDockerfileTemplate, "Dockerfile.tmpl")
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetDockerignore, ".dockerignore")
	if err != nil {
		return err
	}

	return p.CopyTemplate(p.Templates.GetGitignore, ".gitignore")
}
