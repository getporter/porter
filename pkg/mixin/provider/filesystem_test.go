package mixinprovider

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystem_GetMixins(t *testing.T) {
	// Do this in a temp directory so that we can control which mixins show up in the list
	c := config.NewTestConfig(t)
	c.SetHomeDir(os.TempDir())
	c.FileSystem = &afero.Afero{Fs: afero.NewOsFs()} // Hit the real file system for this test

	mixinsDir, err := c.GetMixinsDir()
	require.Nil(t, err)

	// Just copy in the exec and helm mixins
	srcMixinsDir := filepath.Join(c.TestContext.FindBinDir(), "mixins")
	c.CopyDirectory(filepath.Join(srcMixinsDir, "helm"), mixinsDir, true)
	c.CopyDirectory(filepath.Join(srcMixinsDir, "exec"), mixinsDir, true)

	p := NewFileSystem(c.Config)
	mixins, err := p.List()

	require.Nil(t, err)
	require.Len(t, mixins, 2)
	assert.Equal(t, mixins[0].Name, "exec")
	assert.Equal(t, mixins[1].Name, "helm")

	dir, err := os.Stat(mixins[0].Dir)
	require.NoError(t, err)
	assert.True(t, dir.IsDir())
	assert.Equal(t, dir.Name(), "exec")

	binary, err := os.Stat(mixins[0].ClientPath)
	require.NoError(t, err)
	assert.True(t, binary.Mode().IsRegular())
	assert.Equal(t, binary.Name(), "exec")
}

func TestFileSystem_GetSchema(t *testing.T) {
	c := config.NewTestConfig(t)
	// Hit the real file system for this test
	c.FileSystem = &afero.Afero{Fs: afero.NewOsFs()}
	c.NewCommand = exec.Command

	// bin is my home now
	binDir := c.TestContext.FindBinDir()
	c.SetHomeDir(binDir)

	p := NewFileSystem(c.Config)
	mixins, err := p.List()
	require.NoError(t, err)

	var e *mixin.Metadata
	for _, m := range mixins {
		if m.Name == "exec" {
			e = &m
			break
		}
	}
	require.NotNil(t, e)

	gotSchema, err := p.GetSchema(*e)
	require.NoError(t, err)

	wantSchema, err := ioutil.ReadFile("../../exec/testdata/schema.json")
	require.NoError(t, err)

	assert.Equal(t, string(wantSchema), string(gotSchema))
}
