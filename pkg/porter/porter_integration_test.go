//go:build integration

package porter

import (
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/require"
)

func TestPorter_FixPermissions(t *testing.T) {
	p := NewTestPorter(t)
	ctx := p.SetupIntegrationTest()
	defer p.Close()

	home, _ := p.GetHomeDir()
	p.ConfigFilePath = filepath.Join(home, "config.toml")

	testcases := []string{
		filepath.Join(home, "config.toml"),
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			dir := filepath.Dir(tc)
			require.NoError(t, os.MkdirAll(dir, pkg.FileModeDirectory))
			require.NoError(t, os.WriteFile(tc, []byte(""), 0750))

			err := p.FixPermissions(ctx)
			require.NoError(t, err)

			// Check that all files in the directory have the correct permissions
			tests.AssertDirectoryPermissionsEqual(t, dir, pkg.FileModeWritable)
		})
	}
}

func TestPorter_FixPermissions_NoConfigFile(t *testing.T) {
	p := NewTestPorter(t)
	ctx := p.SetupIntegrationTest()
	defer p.Close()

	// Remember the original permissions on the current working directory
	wd := p.Getwd()
	wdInfo, err := p.FileSystem.Stat(wd)
	require.NoError(t, err, "stat on the current working directory failed")
	wantMode := wdInfo.Mode()

	err = p.FixPermissions(ctx)
	require.NoError(t, err)

	// Check that the current working directory didn't have its permissions changed
	wdInfo, err = p.FileSystem.Stat(wd)
	require.NoError(t, err, "stat on the current working directory failed")
	gotMode := wdInfo.Mode()
	tests.AssertFilePermissionsEqual(t, wd, wantMode, gotMode)
}
