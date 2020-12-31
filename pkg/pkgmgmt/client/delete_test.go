package client

import (
	"path"
	"testing"

	"get.porter.sh/porter/pkg/pkgmgmt"

	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestFileSystem_Delete_DeletePackage(t *testing.T) {
	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config, "packages")

	opts := pkgmgmt.UninstallOptions{
		Name: "mixxin",
	}

	err := p.Uninstall(opts)

	assert.Nil(t, err)

	// Make sure the package directory was removed
	pkgDir := path.Join(p.GetPackagesDir(), "mixxin")
	dirExists, _ := p.FileSystem.DirExists(pkgDir)
	assert.False(t, dirExists)
}
