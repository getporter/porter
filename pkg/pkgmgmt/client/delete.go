package client

import (
	"fmt"
	"path/filepath"

	"get.porter.sh/porter/pkg/pkgmgmt"
)

func (fs *FileSystem) Uninstall(opts pkgmgmt.UninstallOptions) error {
	if opts.Name != "" {
		return fs.uninstallByName(opts.Name)
	}

	return fmt.Errorf("No %s name was provided to uninstall", fs.PackageType)
}

func (fs *FileSystem) uninstallByName(name string) error {
	parentDir, err := fs.GetPackagesDir()
	if err != nil {
		return err
	}
	pkgDir := filepath.Join(parentDir, name)
	exists, _ := fs.FileSystem.Exists(pkgDir)
	if exists {
		err = fs.FileSystem.RemoveAll(pkgDir)
		if err != nil {
			return fmt.Errorf("could not remove %s directory %q: %w", fs.PackageType, pkgDir, err)
		}

		return nil
	}

	if fs.Debug {
		fmt.Fprintf(fs.Err, "Unable to find requested %s %s\n", fs.PackageType, name)
	}

	return nil
}
