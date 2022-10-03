package client

import (
	"context"
	"path"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/stretchr/testify/assert"
)

func TestFileSystem_Delete_DeletePackage(t *testing.T) {
	ctx := context.Background()
	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config, "packages")

	parentDir, _ := p.GetPackagesDir()
	pkgDir := path.Join(parentDir, "mixxin")

	opts := pkgmgmt.UninstallOptions{
		Name: "mixxin",
	}

	err := p.Uninstall(ctx, opts)

	assert.Nil(t, err)

	// Make sure the package directory was removed
	dirExists, _ := p.FileSystem.DirExists(pkgDir)
	assert.False(t, dirExists)
}
