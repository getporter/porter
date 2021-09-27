// +build integration

package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestStore_CheckPermissions_Fail(t *testing.T) {
	c := config.NewTestConfig(t)
	l := hclog.New(hclog.DefaultOptions)
	c.SetupIntegrationTest()
	defer c.TestContext.Cleanup()

	home, _ := c.GetHomeDir()
	c.ConfigFilePath = filepath.Join(home, "config.toml")

	testcases := []struct {
		name       string
		path       string
		mode       os.FileMode
		shouldPass bool
	}{
		{name: "bad credentials", path: filepath.Join(home, "credentials", "sabotage.txt"), mode: 0750},
		{name: "bad parameters", path: filepath.Join(home, "parameters", "sabotage.txt"), mode: 0750},
		{name: "bad claims", path: filepath.Join(home, "claims", "sabotage.txt"), mode: 0750},
		{name: "bad outputs", path: filepath.Join(home, "outputs", "sabotage.txt"), mode: 0750},
		{name: "bad config", path: filepath.Join(home, "config.toml"), mode: 0750},
		{name: "check succeeds", path: filepath.Join(home, "config.toml"), mode: 0600},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, os.MkdirAll(filepath.Dir(tc.path), 0700))
			require.NoError(t, os.WriteFile(tc.path, []byte(""), tc.mode))

			s := Store{logger: l, Config: *c.Config}
			err := s.CheckFilePermissions()
			if tc.shouldPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
