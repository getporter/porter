//go:build integration
// +build integration

package porter

import (
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/require"
)

func TestPorter_FixPermissions(t *testing.T) {
	p := NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()

	home, _ := p.GetHomeDir()
	p.ConfigFilePath = filepath.Join(home, "config.toml")

	testcases := []string{
		filepath.Join(home, "credentials", "sabotage.txt"),
		filepath.Join(home, "parameters", "sabotage.txt"),
		filepath.Join(home, "claims", "sabotage.txt"),
		filepath.Join(home, "outputs", "sabotage.txt"),
		filepath.Join(home, "config.toml"),
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			dir := filepath.Dir(tc)
			require.NoError(t, os.MkdirAll(dir, 0700))
			require.NoError(t, os.WriteFile(tc, []byte(""), 0750))

			err := p.FixPermissions()
			require.NoError(t, err)

			// Check that all files in the directory have the correct permissions
			tests.AssertDirectoryPermissionsEqual(t, dir, 0600)
		})
	}
}

func TestPorter_FixPermissions_NoConfigFile(t *testing.T) {
	p := NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()

	// Remember the original permissions on the current working directory
	wd := p.Getwd()
	wdInfo, err := p.FileSystem.Stat(wd)
	require.NoError(t, err, "stat on the current working directory failed")
	wantMode := wdInfo.Mode()

	err = p.FixPermissions()
	require.NoError(t, err)

	// Check that the current working directory didn't have its permissions changed
	wdInfo, err = p.FileSystem.Stat(wd)
	require.NoError(t, err, "stat on the current working directory failed")
	gotMode := wdInfo.Mode()
	tests.AssertFilePermissionsEqual(t, wd, wantMode, gotMode)
}
