package mixinprovider

import (
	"path"
	"testing"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/stretchr/testify/assert"
)

func TestFileSystem_Delete_DeleteMixin(t *testing.T) {
	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config)

	mixinsDir, _ := p.GetMixinsDir()
	mixinDir := path.Join(mixinsDir, "mixxin")

	opts := mixin.DeleteOptions{
		Name: "mixxin",
	}

	_, err := p.Delete(opts)

	assert.Nil(t, err)

	// Make sure the mixin directory was removed
	mixinDirExists, _ := p.FileSystem.DirExists(mixinDir)
	assert.False(t, mixinDirExists)
}
