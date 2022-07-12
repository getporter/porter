package client

import (
	"context"
	"fmt"
	"path/filepath"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/tracing"
)

func (fs *FileSystem) Uninstall(ctx context.Context, opts pkgmgmt.UninstallOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if opts.Name != "" {
		return fs.uninstallByName(ctx, opts.Name)
	}

	return span.Error(fmt.Errorf("No %s name was provided to uninstall", fs.PackageType))
}

func (fs *FileSystem) uninstallByName(ctx context.Context, name string) error {
	log := tracing.LoggerFromContext(ctx)

	parentDir, err := fs.GetPackagesDir()
	if err != nil {
		return log.Error(err)
	}
	pkgDir := filepath.Join(parentDir, name)
	exists, _ := fs.FileSystem.Exists(pkgDir)
	if exists {
		err = fs.FileSystem.RemoveAll(pkgDir)
		if err != nil {
			return log.Error(fmt.Errorf("could not remove %s directory %q: %w", fs.PackageType, pkgDir, err))
		}

		return nil
	}

	log.Debugf("Unable to find requested %s %s\n", fs.PackageType, name)

	return nil
}
