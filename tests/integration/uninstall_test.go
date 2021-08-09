// +build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
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
			wantError:    "Installation not found",
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
			wantError:           porter.ErrUnsafeInstallationDeleteRetryForceDelete.Error(),
		}, {
			name:           "uninstall fails; --force-delete",
			uninstallFails: true,
			forceDelete:    true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := porter.NewTestPorter(t)
			defer p.Teardown()
			p.SetupIntegrationTest()
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
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantError)
			} else {
				require.NoError(t, err, "UninstallBundle failed")
			}

			_, err = p.Claims.GetInstallation(opts.Namespace, opts.Name)
			if tc.installationRemains {
				require.NoError(t, err, "Installation is expected to exist")
			} else {
				require.ErrorIs(t, err, storage.ErrNotFound{})
			}
		})
	}
}
