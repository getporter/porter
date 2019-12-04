package mixinprovider

import (
	"path"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/mixin"
	"github.com/stretchr/testify/assert"
)

func TestFileSystem_Delete_DeleteMixin(t *testing.T) {
	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config)

	mixinsDir, _ := p.GetMixinsDir()
	mixinDir := path.Join(mixinsDir, "mixxin")

	opts := mixin.UninstallOptions{
		Name: "mixxin",
	}

	_, err := p.Uninstall(opts)

	assert.Nil(t, err)

	// Make sure the mixin directory was removed
	mixinDirExists, _ := p.FileSystem.DirExists(mixinDir)
	assert.False(t, mixinDirExists)
}
