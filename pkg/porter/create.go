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

// Create function creates a porter configuration in the specified directory or in the current directory if no directory is specified.
func (p *Porter) Create(bundleName string) error {
	// Normalize the bundleName by removing trailing slashes
	bundleName = strings.TrimSuffix(bundleName, "/")

	// Use the current directory if no directory is passed
	if bundleName == "" {
		bundleName = p.FileSystem.Getwd()
	}

	// Check if the directory in which bundle needs to be created already exists.
	// If not, create the directory.
	_, err := os.Stat(bundleName)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(bundleName, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory for bundle: %w", err)
		}
	}

	err = p.CopyTemplate(p.Templates.GetManifest, filepath.Join(bundleName, config.Name))
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetManifestHelpers, filepath.Join(bundleName, "helpers.sh"))
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetReadme, filepath.Join(bundleName, "README.md"))
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetDockerfileTemplate, filepath.Join(bundleName, "template.Dockerfile"))
	if err != nil {
		return err
	}

	err = p.CopyTemplate(p.Templates.GetDockerignore, filepath.Join(bundleName, ".dockerignore"))
	if err != nil {
		return err
	}

	return p.CopyTemplate(p.Templates.GetGitignore, filepath.Join(bundleName, ".gitignore"))
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
	if _, err := os.Stat(dest); err == nil {
		fmt.Fprintf(os.Stderr, "WARNING: File %q already exists. Overwriting.\n", dest)
	}
	err = p.FileSystem.WriteFile(dest, tmpl, mode)
	if err != nil {
		return fmt.Errorf("failed to write template to %s: %w", dest, err)
	}
	return nil
}
