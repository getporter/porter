package client

import (
	"fmt"
	"path/filepath"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/pkg/errors"
)

func (fs *FileSystem) Uninstall(opts pkgmgmt.UninstallOptions) error {
	if opts.Name != "" {
		return fs.uninstallByName(opts.Name)
	}

	return errors.Errorf("No %s name was provided to uninstall", fs.PackageType)
}

func (fs *FileSystem) uninstallByName(name string) error {
	pkgDir := filepath.Join(fs.GetPackagesDir(), name)
	exists, _ := fs.FileSystem.Exists(pkgDir)
	if exists == true {
		err := fs.FileSystem.RemoveAll(pkgDir)
		if err != nil {
			return errors.Wrapf(err, "could not remove %s directory %q", fs.PackageType, pkgDir)
		}

		return nil
	}

	if fs.Debug {
		fmt.Fprintf(fs.Err, "Unable to find requested %s %s\n", fs.PackageType, name)
	}

	return nil
}
