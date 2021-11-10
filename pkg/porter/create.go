package porter

import (
	"fmt"
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg/config"
	"github.com/pkg/errors"
)

func (p *Porter) Create() error {
	fmt.Fprintln(p.Out, "creating porter configuration in the current directory")

	err := p.CopyTemplate(p.Templates.GetManifest, config.Name)
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetManifestHelpers, "helpers.sh")
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetReadme, "README.md")
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetDockerfileTemplate, "template.Dockerfile")
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetDockerignore, ".dockerignore")
	if err != nil {
		return err
	}

	return p.CopyTemplate(p.Templates.GetGitignore, ".gitignore")
}

func (p *Porter) CopyTemplate(getTemplate func() ([]byte, error), dest string) error {
	tmpl, err := getTemplate()
	if err != nil {
		return err
	}

	var mode os.FileMode = 0600
	if filepath.Ext(dest) == ".sh" {
		mode = 0700
	}

	err = p.FileSystem.WriteFile(dest, tmpl, mode)
	if err != nil {
		return errors.Wrapf(err, "failed to write template to %s", dest)
	}
	return nil
}
