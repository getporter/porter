package porter

import (
	"path/filepath"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_IsInstallationInSync(t *testing.T) {
	cxt := portercontext.New()
	bun, err := cnab.LoadBundle(cxt, filepath.Join("testdata/bundle.json"))
	require.NoError(t, err)

	t.Run("new installation with uninstalled true", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.Installation{
			Uninstalled: true,
		}
		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, NewInstallOptions())
		require.NoError(t, err)
		assert.True(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Ignoring because installation.uninstalled is true but the installation doesn't exist yet")
	})

	t.Run("new installation with uninstalled false", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.Installation{}
		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, NewInstallOptions())
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the installation has not completed successfully yet")
	})

	t.Run("installed - no changes", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.Installation{
			Status: storage.InstallationStatus{
				Installed: &now,
			},
		}
		run := storage.Run{
			// Use the default values from the bundle.json so that we don't trigger reconciliation
			Parameters: storage.NewInternalParameterSet(i.Namespace, i.Name, storage.ValueStrategy("my-second-param", "spring-music-demo")),
		}
		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun}
		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.True(t, insync)
		// Nothing is printed out in this case, the calling function will print "up-to-date" for us
	})

	t.Run("installed - bundle digest changed", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.Installation{
			Status: storage.InstallationStatus{
				Installed:    &now,
				BundleDigest: "olddigest",
			},
		}
		run := storage.Run{
			BundleDigest: "olddigest",
		}
		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun, Digest: "newdigest"}
		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the bundle definition has changed")
	})

	t.Run("installed - param changed", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.Installation{
			Status: storage.InstallationStatus{
				Installed: &now,
			},
		}
		run := storage.Run{
			Parameters: storage.NewInternalParameterSet(i.Namespace, i.Name, storage.ValueStrategy("my-second-param", "newvalue")),
		}
		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun}
		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the parameters have changed")

	})

	t.Run("installed - credential set changed", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.Installation{
			CredentialSets: []string{"newcreds"},
			Status: storage.InstallationStatus{
				Installed: &now,
			},
		}
		run := storage.Run{
			CredentialSets: []string{"oldcreds"},
			// Use the default values from the bundle.json so they don't trigger the reconciliation
			Parameters: storage.NewInternalParameterSet(i.Namespace, i.Name, storage.ValueStrategy("my-second-param", "spring-music-demo")),
		}
		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun}
		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the credential set names have changed")

	})

	t.Run("installed - uninstalled change to true", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.Installation{
			Uninstalled: true, // trigger uninstall
			Status: storage.InstallationStatus{
				Installed: &now,
			},
		}
		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, NewUninstallOptions())
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because installation.uninstalled is true")
	})

	t.Run("uninstalled: uninstalled set to back to false", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		installTime := now.Add(-time.Second * 5)
		i := storage.Installation{
			Uninstalled: false,
			Status: storage.InstallationStatus{
				Installed:   &installTime,
				Uninstalled: &now,
			},
		}
		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, NewUninstallOptions())
		require.NoError(t, err)
		assert.True(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Ignoring because the installation is uninstalled")
	})
}
