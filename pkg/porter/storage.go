package porter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func (p *Porter) MigrateStorage() error {
	logfilePath, err := p.Storage.Migrate()

	fmt.Fprintf(p.Out, "\nSaved migration logs to %s\n", logfilePath)

	if err != nil {
		// The error has already been printed, don't return it otherwise it will be double printed
		return errors.New("Migration failed!")
	}

	fmt.Fprintln(p.Out, "Migration complete!")
	return nil
}

func (p *Porter) FixPermissions() error {
	home, _ := p.GetHomeDir()
	fmt.Fprintf(p.Out, "Resetting file permissions in %s...\n", home)

	// Fix as many files as we can, and then report any errors
	fixFile := func(path string, mode os.FileMode) error {
		info, err := p.FileSystem.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			} else {
				return errors.Wrapf(err, "error checking file permissions for %s", path)
			}
		}

		gotPerms := info.Mode().Perm()
		if mode != gotPerms|mode {
			if err := p.FileSystem.Chmod(path, mode); err != nil {
				return errors.Wrapf(err, "could not set permissions on file %s to %o", path, mode)
			}
		}
		return nil
	}

	fixDir := func(dir string, mode os.FileMode) error {
		var bigErr *multierror.Error
		p.FileSystem.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				if !os.IsNotExist(err) {
					bigErr = multierror.Append(bigErr, errors.Wrapf(err, "error walking path %s", path))
				}
				return nil
			}

			if info.IsDir() {
				if err := p.FileSystem.Chmod(path, pkg.FileModeDirectory); err != nil {
					bigErr = multierror.Append(bigErr, errors.Wrapf(err, "could not set permissions on directory %s to %o", path, mode))
				}
			} else {
				if err = fixFile(path, mode); err != nil {
					bigErr = multierror.Append(bigErr, err)
				}
			}
			return nil
		})
		return bigErr.ErrorOrNil()
	}

	var bigErr *multierror.Error
	dataFiles := []string{p.ConfigFilePath, filepath.Join(home, "schema.json")}
	for _, file := range dataFiles {
		if err := fixFile(file, pkg.FileModeWritable); err != nil {
			bigErr = multierror.Append(bigErr, err)
		}
	}

	dataDirs := []string{"installations", "claims", "results", "outputs", "cache", "credentials", "parameters"}
	for _, dir := range dataDirs {
		if err := fixDir(filepath.Join(home, dir), pkg.FileModeWritable); err != nil {
			bigErr = multierror.Append(bigErr, err)
		}
	}

	porterPath, _ := p.GetPorterPath()
	binFiles := []string{porterPath}
	for _, file := range binFiles {
		if err := fixFile(file, pkg.FileModeExecutable); err != nil {
			bigErr = multierror.Append(bigErr, err)
		}
	}

	binDirs := []string{"mixins", "plugins", "runtimes"}
	for _, dir := range binDirs {
		if err := fixDir(filepath.Join(home, dir), pkg.FileModeExecutable); err != nil {
			bigErr = multierror.Append(bigErr, err)
		}
	}

	return bigErr.ErrorOrNil()
}
