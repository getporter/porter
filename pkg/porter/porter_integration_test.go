//go:build integration
// +build integration

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
	p.SetupIntegrationTest()
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

			err := p.FixPermissions()
			require.NoError(t, err)

			tests.AssertDirectoryPermissionsEqual(t, dir, pkg.FileModeWritable)
		})
	}
}
