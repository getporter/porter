package porter

import (
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
	if err != nil {
		if os.IsNotExist(err) {
			// This code here attempts to create the directory in which bundle needs to be created,
			// if the directory does not exist.
			// bundleName can handle both the relative path and absolute path into consideration,
			// For example if we want to create a bundle named mybundle in an existing directory /home/user we can call porter create /home/user/mybundle or porter create mybundle in the /home/user directory.
			// If we are in a directory /home/user and we want to create mybundle in the directory /home/user/directory given the directory exists,
			// we can call porter create directory/mybundle from the /home/user directory or with any relative paths' combinations that one can come up with.
			// Only condition to use porter create with absolute and relative paths is that all the directories in the path except the last one should strictly exist.
			err = os.Mkdir(bundleName, os.ModePerm)
			// This error message is returned when the os.Mkdir call encounters an error
			// during the directory creation process. It specifically indicates that the attempt
			// to create the bundle directory failed. This could occur due to reasons such as
			// lack of permissions, a file system error, or if the parent directory doesn't exist.
			if err != nil {
				return fmt.Errorf("failed to create directory for bundle: %w", err)
			}
			// This error message is returned when the os.Stat call encounters an error other than
			// the directory not existing. It implies that there was an issue with checking the bundle directory,
			// but it doesn't mean that the directory creation itself failed.
			// It could be due to various reasons, such as insufficient permissions, an invalid directory path,
			// or other file system-related errors.
		} else {
			return fmt.Errorf("failed to check bundle directory: %w", err)
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

	fmt.Fprintf(p.Out, "creating porter configuration in %s\n", bundleName)

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

	err = p.FileSystem.WriteFile(dest, tmpl, mode)
	if err != nil {
		return fmt.Errorf("failed to write template to %s: %w", dest, err)
	}
	return nil
}
