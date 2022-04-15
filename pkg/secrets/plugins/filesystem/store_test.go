package filesystem_test

import (
	"context"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins/filesystem"
	"github.com/stretchr/testify/require"
)

func TestFileSystem_Permission(t *testing.T) {
	c := config.NewTestConfig(t)
	defer c.Teardown()

	testStore := filesystem.NewStore(c.Config)
	defer testStore.Close()

	ctx := context.Background()
	err := testStore.Connect(ctx)
	require.NoError(t, err)

	secretDir := "/home/myuser/.porter/secrets"
	info, err := c.FileSystem.Stat(secretDir)
	require.NoError(t, err)
	require.Equal(t, filesystem.FileModeSensitiveDirectory, info.Mode().Perm(), "invalid folder permission %s", info.Mode().Perm().String())

	secretKey := "porter-filesystem-plugin-test"
	err = testStore.Create(ctx, secrets.SourceSecret, secretKey, "supersecret")
	require.NoError(t, err)

	info, err = c.FileSystem.Stat(filepath.Join(secretDir, secretKey))
	require.NoError(t, err)
	require.Equal(t, filesystem.FileModeSensitiveWritable, info.Mode().Perm())
}

func TestFileSystem_SetSecretDir(t *testing.T) {
	c := config.NewTestConfig(t)

	s := filesystem.NewStore(c.Config)
	secretDir, err := s.SetSecretDir()
	require.NoError(t, err)
	require.Equal(t, "/home/myuser/.porter/secrets", secretDir)
}

func TestFileSystem_DataOperation(t *testing.T) {
	c := config.NewTestConfig(t)
	defer c.Teardown()

	testStore := filesystem.NewStore(c.Config)
	defer testStore.Close()

	ctx := context.Background()
	err := testStore.Connect(ctx)
	require.NoError(t, err)

	secretKey := "porter-filesystem-plugin-test"
	secretValue := "supersecret"
	err = testStore.Create(ctx, secrets.SourceSecret, secretKey, secretValue)
	require.NoError(t, err)

	data, err := testStore.Resolve(ctx, secrets.SourceSecret, secretKey)
	require.NoError(t, err)
	require.Equal(t, secretValue, data)
}
