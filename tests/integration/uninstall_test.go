// +build integration

package integration

import (
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/require"
)

func TestUninstall_DeleteInstallation(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name                string
		notInstalled        bool
		uninstallFails      bool
		delete              bool
		forceDelete         bool
		installationRemains bool
		wantError           string
	}{
		{
			name:         "not yet installed",
			notInstalled: true,
			wantError:    "1 error occurred:\n\t* could not load installation porter-hello: Installation does not exist\n\n",
		}, {
			name:                "no --delete",
			installationRemains: true,
		}, {
			name:      "--delete",
			delete:    true,
			wantError: "",
		}, {
			name:                "uninstall fails; --delete",
			uninstallFails:      true,
			delete:              true,
			installationRemains: true,
			wantError:           fmt.Sprintf("2 errors occurred:\n\t* Command driver (uninstall) failed executing bundle: exit status 1\n\t* %s\n\n", porter.ErrUnsafeInstallationDeleteRetryForceDelete.Error()),
		}, {
			name:           "uninstall fails; --force-delete",
			uninstallFails: true,
			forceDelete:    true,
		},
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
			if !tc.notInstalled {
				opts := porter.NewInstallOptions()
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

			path := p.Getenv("PATH")
			dir := p.AddTestDriver(driver)
			defer p.Setenv("PATH", path)
			defer p.TestConfig.TestContext.FileSystem.RemoveAll(dir)

			// Uninstall bundle with custom command driver
			opts := porter.NewUninstallOptions()
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
			if tc.installationRemains {
				require.NoError(t, err, "Installation is expected to exist")
			} else {
				require.EqualError(t, err, "Installation does not exist")
			}
		})
	}
}
