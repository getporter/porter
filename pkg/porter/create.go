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

// Create creates a new bundle configuration with the specified bundleName. A directory with the given bundleName will be created if it does not already exist.
// If bundleName is the empty string, the configuration will be created in the current directory, and the name will be "porter-hello".
func (p *Porter) Create(bundleName string) error {
	// Normalize the bundleName by removing trailing slashes
	bundleName = strings.TrimSuffix(bundleName, "/")

	// If given a bundleName, create directory if it doesn't exist
	if bundleName != "" {
		_, err := os.Stat(bundleName)
		if err != nil && errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(bundleName, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create directory for bundle: %w", err)
			}
		}
	}

	var err error
	if bundleName == "" {
		// create bundle with default name "porter_hello"
		err = p.CopyTemplate(p.Templates.GetManifest, filepath.Join(bundleName, config.Name))
	} else {
		// create bundle with given name
		err = p.CopyTemplate(func() ([]byte, error) {
			content, err := p.Templates.GetManifest()
			if err != nil {
				return nil, err
			}
			content = []byte(strings.ReplaceAll(string(content), "porter-hello", bundleName))
			return content, nil
		}, filepath.Join(bundleName, config.Name))
	}
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
