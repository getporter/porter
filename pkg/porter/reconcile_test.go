package porter

import (
	"context"
	"get.porter.sh/porter/pkg/secrets"
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
	const helloRef = "ghcr.io/getporter/examples/porter-hello:v0.2.0"

	cxt := portercontext.New()
	bun, err := cnab.LoadBundle(cxt, filepath.Join("testdata/bundle.json"))
	require.NoError(t, err)

	t.Run("new installation with uninstalled true", func(t *testing.T) {
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.NewInstallation("", "mybuns")
		i.Uninstalled = true
		opts := NewInstallOptions()
		opts.Reference = helloRef
		require.NoError(t, p.applyActionOptionsToInstallation(ctx, opts, &i))

		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, opts)
		require.NoError(t, err)
		assert.True(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Ignoring because installation.uninstalled is true but the installation doesn't exist yet")
	})

	t.Run("new installation with uninstalled false", func(t *testing.T) {
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.NewInstallation("", "mybuns")
		opts := NewInstallOptions()
		opts.Reference = helloRef
		require.NoError(t, p.applyActionOptionsToInstallation(ctx, opts, &i))

		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, opts)
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the installation has not completed successfully yet")
	})

	t.Run("installed - no changes", func(t *testing.T) {
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		myps := storage.NewParameterSet("", "myps")
		myps.SetStrategy("my-second-param", secrets.HardCodedValueStrategy("override"))

		err := p.Parameters.InsertParameterSet(ctx, myps)
		require.NoError(t, err)

		i := storage.NewInstallation("", "mybuns")
		i.ParameterSets = []string{"myps"}
		i.Status.Installed = &now
		run := i.NewRun(cnab.ActionInstall, bun)
		run.Parameters.SetStrategy("my-second-param", secrets.HardCodedValueStrategy("override"))

		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun}
		require.NoError(t, p.applyActionOptionsToInstallation(ctx, upgradeOpts, &i))

		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.True(t, insync)
		// Nothing is printed out in this case, the calling function will print "up-to-date" for us
	})

	t.Run("installed - bundle digest changed", func(t *testing.T) {
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.NewInstallation("", "mybuns")
		i.Status.Installed = &now
		i.Status.BundleDigest = "olddigest"
		run := storage.Run{
			BundleDigest: "olddigest",
		}
		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun, Digest: "newdigest"}
		require.NoError(t, p.applyActionOptionsToInstallation(ctx, upgradeOpts, &i))

		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the bundle definition has changed")
	})

	t.Run("installed - param changed", func(t *testing.T) {
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.NewInstallation("", "mybuns")
		i.Status.Installed = &now
		run := i.NewRun(cnab.ActionInstall, bun)
		run.Parameters.SetStrategy("my-second-param", secrets.HardCodedValueStrategy("new value"))

		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun}
		require.NoError(t, p.applyActionOptionsToInstallation(ctx, upgradeOpts, &i))

		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the parameters have changed")

	})

	t.Run("installed - credential set changed", func(t *testing.T) {
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.NewInstallation("", "mybuns")
		i.Status.Installed = &now
		i.CredentialSets = []string{"newcreds"}
		run := i.NewRun(cnab.ActionInstall, bun)
		run.CredentialSets = []string{"oldcreds"}
		run.Parameters.SetStrategy("my-second-param", secrets.HardCodedValueStrategy("spring-music-demo"))

		upgradeOpts := NewUpgradeOptions()
		upgradeOpts.bundleRef = &cnab.BundleReference{Definition: bun}
		require.NoError(t, p.applyActionOptionsToInstallation(ctx, upgradeOpts, &i))

		insync, err := p.IsInstallationInSync(p.RootContext, i, &run, upgradeOpts)
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because the credential set names have changed")

	})

	t.Run("installed - uninstalled change to true", func(t *testing.T) {
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		i := storage.NewInstallation("", "mybuns")
		i.Uninstalled = true // trigger uninstall
		i.Status.Installed = &now
		opts := NewUninstallOptions()
		opts.Reference = helloRef
		require.NoError(t, p.applyActionOptionsToInstallation(ctx, opts, &i))

		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, opts)
		require.NoError(t, err)
		assert.False(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Triggering because installation.uninstalled is true")
	})

	t.Run("uninstalled: uninstalled set to back to false", func(t *testing.T) {
		ctx := context.Background()
		p := NewTestPorter(t)
		defer p.Close()

		installTime := now.Add(-time.Second * 5)
		i := storage.NewInstallation("", "mybuns")
		i.Uninstalled = false
		i.Status.Installed = &installTime
		i.Status.Uninstalled = &now
		opts := NewUninstallOptions()
		opts.Reference = helloRef
		require.NoError(t, p.applyActionOptionsToInstallation(ctx, opts, &i))

		insync, err := p.IsInstallationInSync(p.RootContext, i, nil, NewUninstallOptions())
		require.NoError(t, err)
		assert.True(t, insync)
		assert.Contains(t, p.TestConfig.TestContext.GetError(), "Ignoring because the installation is uninstalled")
	})
}
