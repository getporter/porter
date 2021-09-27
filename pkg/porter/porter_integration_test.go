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

			tests.AssertDirectoryPermissionsEqual(t, dir, 0600)
		})
	}
}
