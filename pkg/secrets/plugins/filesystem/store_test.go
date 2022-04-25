package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins/filesystem"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestFileSystem_Permission(t *testing.T) {
	c := config.NewTestConfig(t)
	ctx := c.TestContext
	defer ctx.Teardown()
	_, homeDir := ctx.UseFilesystem()
	c.SetHomeDir(homeDir)

	cfg := filesystem.NewConfig(ctx.DebugPlugins, homeDir)
	secretDir, err := cfg.SetSecretDir()
	require.NoError(t, err)
	testStore := filesystem.NewStore(cfg, hclog.NewNullLogger(), ctx.FileSystem)
	defer testStore.Close()

	err = testStore.Connect()
	require.NoError(t, err)

	info, err := ctx.FileSystem.Stat(secretDir)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(filesystem.FileModeSensitiveDirectory), info.Mode().Perm(), "invalid folder permission %s", info.Mode().Perm().String())

	secretKey := "porter-filesystem-plugin-test"
	err = testStore.Create(secrets.SourceSecret, secretKey, "supersecret")
	require.NoError(t, err)

	info, err = ctx.FileSystem.Stat(filepath.Join(secretDir, secretKey))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(filesystem.FileModeSensitiveWritable), info.Mode().Perm())
}

func TestFileSystem_Config_SetSecretDir(t *testing.T) {
	c := config.NewTestConfig(t)

	c.SetHomeDir("")
	home, err := c.GetHomeDir()
	require.NoError(t, err)

	cfg := filesystem.NewConfig(c.TestContext.DebugPlugins, home)
	secretDir, err := cfg.SetSecretDir()
	require.NoError(t, err)
	require.True(t, cfg.Valid())
	require.Equal(t, filepath.Join(home, filesystem.SECRET_FOLDER), secretDir)
}

func TestFileSystem_DataOperation(t *testing.T) {
	c := config.NewTestConfig(t)
	ctx := c.TestContext
	defer ctx.Teardown()
	_, homeDir := ctx.UseFilesystem()
	c.SetHomeDir(homeDir)

	cfg := filesystem.NewConfig(c.TestContext.DebugPlugins, homeDir)
	_, err := cfg.SetSecretDir()
	require.NoError(t, err)
	testStore := filesystem.NewStore(cfg, hclog.NewNullLogger(), ctx.FileSystem)
	defer testStore.Close()

	err = testStore.Connect()
	require.NoError(t, err)

	secretKey := "porter-filesystem-plugin-test"
	secretValue := "supersecret"
	err = testStore.Create(secrets.SourceSecret, secretKey, secretValue)
	require.NoError(t, err)

	data, err := testStore.Resolve(secrets.SourceSecret, secretKey)
	require.NoError(t, err)
	require.Equal(t, secretValue, data)
}
