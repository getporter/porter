// +build integration

package tests

import (
	"os"
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/require"
)

func TestUninstall_DeleteInstallation(t *testing.T) {
	testcases := []struct {
		name               string
		installed          bool
		uninstallFails     bool
		delete             bool
		forceDelete        bool
		installationExists bool
		wantError          string
	}{
		{"not yet installed", false, false, false, false, false, "1 error occurred:\n\t* could not load installation HELLO: Installation does not exist\n\n"},
		{"no delete", true, false, false, false, true, ""},
		{"delete", true, false, true, false, false, ""},
		{"uninstall fails - delete", true, true, true, false, true, "2 errors occurred:\n\t* Command driver (uninstall) failed executing bundle: exit status 1\n\t* not deleting installation HELLO as uninstall was not successful; use --force-delete to override\n\n"},
		{"uninstall fails - force-delete", true, true, false, true, false, ""},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := porter.NewTestPorter(t)

			p.SetupIntegrationTest()
			defer p.CleanupIntegrationTest()
			p.Debug = false

			err := p.Create()
			require.NoError(t, err)

			// Install bundle
			if tc.installed {
				opts := porter.InstallOptions{}
				opts.Driver = "debug"

				err = opts.Validate(nil, p.Porter)
				require.NoError(t, err, "Validate install options failed")

				err = p.InstallBundle(opts)
				require.NoError(t, err, "InstallBundle failed")
			}

			// Set up command driver
			driver := porter.TestDriver{
				Name:     "uninstall",
				Filepath: "testdata/drivers/exit-driver.sh",
			}
			if tc.uninstallFails {
				driver.Filepath = "testdata/drivers/exit-driver-fail.sh"
			}

			path := os.Getenv("PATH")
			dir := p.AddTestDriver(driver)
			defer os.Setenv("PATH", path)
			defer p.TestConfig.TestContext.FileSystem.RemoveAll(dir)

			// Uninstall bundle with custom command driver
			opts := porter.UninstallOptions{}
			opts.Delete = tc.delete
			opts.ForceDelete = tc.forceDelete
			opts.Driver = driver.Name

			err = opts.Validate(nil, p.Porter)
			require.NoError(t, err, "Validate uninstall options failed")

			err = p.UninstallBundle(opts)
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err, "UninstallBundle failed")
			}

			_, err = p.Claims.ReadInstallation(opts.Name)
			if tc.installationExists {
				require.NoError(t, err, "Installation is expected to exist")
			} else {
				require.EqualError(t, err, "Installation does not exist")
			}
		})
	}
}
