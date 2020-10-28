// +build integration

package mixin

import (
	"io/ioutil"
	"os/exec"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageManager_GetSchema(t *testing.T) {
	c := config.NewTestConfig(t)
	// Hit the real file system for this test
	c.FileSystem = &afero.Afero{Fs: afero.NewOsFs()}
	c.NewCommand = exec.Command

	// bin is my home now
	binDir := c.TestContext.FindBinDir()
	c.SetHomeDir(binDir)

	p := NewPackageManager(c.Config)
	gotSchema, err := p.GetSchema("exec")
	require.NoError(t, err)

	wantSchema, err := ioutil.ReadFile("../exec/schema/exec.json")
	require.NoError(t, err)
	assert.Equal(t, string(wantSchema), gotSchema)
}
