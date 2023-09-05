package porter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/config"
)

// Create creates a new bundle configuration in the current directory
func (p *Porter) Create() error {
	destinationDir := "." // current directory

	if err := p.CopyTemplate(p.Templates.GetManifest, filepath.Join(destinationDir, config.Name)); err != nil {
		return err
	}
	return p.copyAllTemplatesExceptPorterYaml(destinationDir)
}

// CreateInDir creates a new bundle configuration in the specified directory. The directory will be created if it
// doesn't already exist. For example, if dir is "foo/bar/baz", the directory structure "foo/bar/baz" will be created.
// The bundle name will be set to the "base" of the given directory, which is "baz" in the example above.
func (p *Porter) CreateInDir(dir string) error {
	bundleName := filepath.Base(dir)

	// Create dirs if they don't exist
	_, err := p.FileSystem.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		err = p.FileSystem.MkdirAll(dir, 0755)
	}
	if err != nil {
		// the Stat failed with an error different from os.ErrNotExist OR the MkdirAll failed to create the dir(s)
		return fmt.Errorf("failed to create directory for bundle: %w", err)
	}

	// create porter.yaml, using base of given dir as the bundle name
	err = p.CopyTemplate(func() ([]byte, error) {
		content, err := p.Templates.GetManifest()
		if err != nil {
			return nil, err
		}
		content = []byte(strings.ReplaceAll(string(content), "porter-hello", bundleName))
		return content, nil
	}, filepath.Join(dir, config.Name))
	if err != nil {
		return err
	}

	return p.copyAllTemplatesExceptPorterYaml(dir)
}

func (p *Porter) copyAllTemplatesExceptPorterYaml(destinationDir string) error {
	err := p.CopyTemplate(p.Templates.GetManifestHelpers, filepath.Join(destinationDir, "helpers.sh"))
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetReadme, filepath.Join(destinationDir, "README.md"))
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetDockerfileTemplate, filepath.Join(destinationDir, "template.Dockerfile"))
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetDockerignore, filepath.Join(destinationDir, ".dockerignore"))
	if err != nil {
		return err
	}

	return p.CopyTemplate(p.Templates.GetGitignore, filepath.Join(destinationDir, ".gitignore"))
}

func (p *Porter) CopyTemplate(getTemplate func() ([]byte, error), dest string) error {
	tmpl, err := getTemplate()
	if err != nil {
		return err
	}

	var mode os.FileMode = pkg.FileModeWritable
	if filepath.Ext(dest) == ".sh" {
		mode = pkg.FileModeExecutable
	}

	if _, err := p.FileSystem.Stat(dest); err == nil {
		fmt.Fprintf(p.Err, "WARNING: File %q already exists. Overwriting.\n", dest)
	}
	err = p.FileSystem.WriteFile(dest, tmpl, mode)
	if err != nil {
		return fmt.Errorf("failed to write template to %s: %w", dest, err)
	}
	return nil
}
