package porter

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	portercontext "get.porter.sh/porter/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_IsInstallationInSync(t *testing.T) {
	cxt := portercontext.New()
	bun, err := cnab.LoadBundle(cxt, filepath.Join("testdata/bundle.json"))
	require.NoError(t, err)

	t.Run("inactive not yet installed", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Teardown()

		i := claims.Installation{
			Active: false,
		}
		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, NewInstallOptions())
		require.NoError(t, err)
		assert.True(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Ignoring because the installation is inactive")
	})

	t.Run("active not yet installed", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Teardown()

		i := claims.Installation{
			Active: true,
		}
		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, NewInstallOptions())
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the installation has not completed successfully yet")
	})

	t.Run("active installed - no changes", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Teardown()

		i := claims.Installation{
			Active: true,
			Status: claims.InstallationStatus{
				Installed: &now,
			},
		}
		run := claims.Run{
			// Use the default values from the bundle.json so that we don't trigger reconciliation
			Parameters: map[string]interface{}{
				"my-second-param": "spring-music-demo",
			},
		}
		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun}
		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.True(t, insync)
		// Nothing is printed out in this case, the calling function will print "up-to-date" for us
	})

	t.Run("active installed - bundle digest changed", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Teardown()

		i := claims.Installation{
			Active: true,
			Status: claims.InstallationStatus{
				Installed:    &now,
				BundleDigest: "olddigest",
			},
		}
		run := claims.Run{
			BundleDigest: "olddigest",
		}
		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun, Digest: "newdigest"}
		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the bundle definition has changed")
	})

	t.Run("active installed - param changed", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Teardown()

		i := claims.Installation{
			Active: true,
			Status: claims.InstallationStatus{
				Installed: &now,
			},
		}
		run := claims.Run{
			Parameters: map[string]interface{}{
				"my-second-param": "newvalue",
			},
		}
		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun}
		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the parameters have changed")

	})

	t.Run("active installed - credential set changed", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Teardown()

		i := claims.Installation{
			Active:         true,
			CredentialSets: []string{"newcreds"},
			Status: claims.InstallationStatus{
				Installed: &now,
			},
		}
		run := claims.Run{
			CredentialSets: []string{"oldcreds"},
			// Use the default values from the bundle.json so they don't trigger the reconciliation
			Parameters: map[string]interface{}{
				"my-second-param": "spring-music-demo",
			},
		}
		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun}
		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the credential set names have changed")

	})

	t.Run("inactive installed", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Teardown()

		i := claims.Installation{
			Active: false,
			Status: claims.InstallationStatus{
				Installed: &now,
			},
		}
		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, NewUninstallOptions())
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the installation is now inactive")
	})

	t.Run("inactive uninstalled", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Teardown()

		i := claims.Installation{
			Active: false,
			Status: claims.InstallationStatus{
				Installed:   &now,
				Uninstalled: &now,
			},
		}
		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, NewUninstallOptions())
		require.NoError(t, err)
		assert.True(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Ignoring because the installation is uninstalled")
	})
}
