package porter

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/hashicorp/go-multierror"
)

type MigrateStorageOptions struct {
	OldHome           string
	OldStorageAccount string
	Namespace         string
}

func (o MigrateStorageOptions) Validate() error {
	if o.OldHome == "" {
		return errors.New("--old-home is required")
	}

	return nil
}

func (p *Porter) MigrateStorage(ctx context.Context, opts MigrateStorageOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	migrateOpts := storage.MigrateOptions{
		OldHome:           opts.OldHome,
		OldStorageAccount: opts.OldStorageAccount,
		NewNamespace:      opts.Namespace,
	}
	return p.Storage.Migrate(ctx, migrateOpts)
}

func (p *Porter) FixPermissions(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	home, err := p.GetHomeDir()
	if err != nil {
		return err
	}

	span.Infof("Resetting file permissions in %s", home)

	// Fix as many files as we can, and then report any errors
	fixFile := func(path string, mode os.FileMode) error {
		info, err := p.FileSystem.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			} else {
				return span.Error(fmt.Errorf("error checking file permissions for %s: %w", path, err))
			}
		}

		if info.IsDir() {
			return span.Error(fmt.Errorf("fixFile was called on a directory %s", path))
		}

		if _, err = filepath.Rel(home, path); err != nil {
			return span.Error(fmt.Errorf("fixFile was called on a path, %s, that isn't in the PORTER_HOME directory %s", path, home))
		}

		gotPerms := info.Mode().Perm()
		if mode != gotPerms|mode {
			if err := p.FileSystem.Chmod(path, mode); err != nil {
				return span.Error(fmt.Errorf("could not set permissions on file %s to %o: %w", path, mode, err))
			}
		}
		return nil
	}

	fixDir := func(dir string, mode os.FileMode) error {
		var bigErr *multierror.Error
		p.FileSystem.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				if !os.IsNotExist(err) {
					bigErr = multierror.Append(bigErr, fmt.Errorf("error walking path %s: %w", path, err))
				}
				return nil
			}

			if info.IsDir() {
				if err := p.FileSystem.Chmod(path, pkg.FileModeDirectory); err != nil {
					bigErr = multierror.Append(bigErr, fmt.Errorf("could not set permissions on directory %s to %o: %w", path, mode, err))
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
	dataFiles := []string{filepath.Join(home, "schema.json")}
	if p.ConfigFilePath != "" {
		dataFiles = append(dataFiles, p.ConfigFilePath)
	}
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

	porterPath, _ := p.GetPorterPath(ctx)
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

	return span.Error(bigErr.ErrorOrNil())
}
