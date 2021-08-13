// +build integration

package porter

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"github.com/stretchr/testify/require"
)

func TestInstallFromTag_ManageFromClaim(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	cacheDir, _ := p.Cache.GetCacheDir()
	p.TestConfig.TestContext.AddTestDirectory("testdata/cache", cacheDir)

	installOpts := NewInstallOptions()
	installOpts.Name = "hello"
	installOpts.Reference = "getporter/porter-hello:v0.1.1"
	err := installOpts.Validate(nil, p.Porter)
	require.NoError(t, err, "InstallOptions.Validate failed")

	err = p.InstallBundle(installOpts)
	require.NoError(t, err, "InstallBundle failed")

	upgradeOpts := NewUpgradeOptions()
	upgradeOpts.Name = installOpts.Name
	err = upgradeOpts.Validate(nil, p.Porter)

	err = p.UpgradeBundle(upgradeOpts)
	require.NoError(t, err, "UpgradeBundle failed")

	uninstallOpts := NewUninstallOptions()
	uninstallOpts.Name = installOpts.Name
	err = uninstallOpts.Validate(nil, p.Porter)

	err = p.UninstallBundle(uninstallOpts)
	require.NoError(t, err, "UninstallBundle failed")
}

func TestResolveBundleReference(t *testing.T) {
	t.Parallel()
	t.Run("current bundle source", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Teardown()

		p.AddTestBundleDir(filepath.Join(p.RepoRoot, "tests/testdata/mybuns"), true)

		opts := &BundleActionOptions{}
		require.NoError(t, opts.Validate(nil, p.Porter))
		ref, err := p.resolveBundleReference(opts)
		require.NoError(t, err)
		require.NotEmpty(t, opts.Name)
		require.NotEmpty(t, ref.Definition)
		require.NotEmpty(t, p.Manifest)
	})

	t.Run("cnab file", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Teardown()

		p.AddTestFile(filepath.Join(p.RepoRoot, "build/testdata/bundles/mysql/.cnab/bundle.json"), "bundle.json")

		opts := &BundleActionOptions{}
		opts.CNABFile = "bundle.json"
		require.NoError(t, opts.Validate(nil, p.Porter))
		ref, err := p.resolveBundleReference(opts)
		require.NoError(t, err)
		require.NotEmpty(t, opts.Name)
		require.NotEmpty(t, ref.Definition)
	})

	t.Run("reference", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Teardown()
		p.SetupIntegrationTest()

		opts := &BundleActionOptions{}
		opts.Reference = "getporter/porter-hello:v0.1.1"
		require.NoError(t, opts.Validate(nil, p.Porter))
		ref, err := p.resolveBundleReference(opts)
		require.NoError(t, err)
		require.NotEmpty(t, opts.Name)
		require.NotEmpty(t, ref.Definition)
		require.NotEmpty(t, ref.RelocationMap)
		require.NotEmpty(t, ref.Digest)
	})

	t.Run("installation name", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Teardown()

		i := p.TestClaims.CreateInstallation(claims.NewInstallation("dev", "example"))
		p.TestClaims.CreateRun(i.NewRun(cnab.ActionInstall), func(r *claims.Run) {
			r.BundleReference = kahnlatest.String()
			r.Bundle = buildExampleBundle()
			r.BundleDigest = kahnlatestHash
		})
		opts := &BundleActionOptions{}
		opts.Name = "example"
		opts.Namespace = "dev"
		require.NoError(t, opts.Validate(nil, p.Porter))
		ref, err := p.resolveBundleReference(opts)
		require.NoError(t, err)
		require.NotEmpty(t, opts.Name)
		require.NotEmpty(t, ref.Definition)
		require.NotEmpty(t, ref.Digest)
	})
}
