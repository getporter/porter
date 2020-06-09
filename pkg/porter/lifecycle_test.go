package porter

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundlePullUpdateOpts_bundleCached(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	home, err := p.TestConfig.GetHomeDir()
	t.Logf("home dir is: %s", home)
	cacheDir, err := p.Cache.GetCacheDir()
	require.NoError(t, err, "should have had a porter cache dir")
	t.Logf("cache dir is: %s", cacheDir)
	p.TestConfig.TestContext.AddTestDirectory("testdata/cache", cacheDir)
	fullPath := filepath.Join(cacheDir, "887e7e65e39277f8744bd00278760b06/cnab/bundle.json")
	fileExists, err := p.FileSystem.Exists(fullPath)
	require.True(t, fileExists, "this test requires that the file exist")
	_, ok, err := p.Cache.FindBundle("deislabs/kubekahn:1.0")
	assert.True(t, ok, "should have found the bundle...")
	b := &BundleLifecycleOpts{
		BundlePullOptions: BundlePullOptions{
			Tag: "deislabs/kubekahn:1.0",
		},
	}
	err = p.prepullBundleByTag(b)
	assert.NoError(t, err, "pulling bundle should not have resulted in an error")
	assert.Equal(t, "mysql", b.Name, "name should have matched testdata bundle")
	assert.Equal(t, fullPath, b.CNABFile, "the prepare method should have set the file to the fullpath")
}

func TestBundlePullUpdateOpts_pullError(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	b := &BundleLifecycleOpts{
		BundlePullOptions: BundlePullOptions{
			Tag: "deislabs/kubekahn:latest",
		},
	}
	err := p.prepullBundleByTag(b)
	assert.Error(t, err, "pulling bundle should have resulted in an error")
	assert.Contains(t, err.Error(), "unable to pull bundle deislabs/kubekahn:latest")

}

func TestBundlePullUpdateOpts_cacheLies(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	// mess up the cache
	p.FileSystem.WriteFile("/root/.porter/cache/887e7e65e39277f8744bd00278760b06/cnab/bundle.json", []byte(""), 0644)

	b := &BundleLifecycleOpts{
		BundlePullOptions: BundlePullOptions{
			Tag: "deislabs/kubekahn:1.0",
		},
	}

	err := p.prepullBundleByTag(b)
	assert.Error(t, err, "pulling bundle should have resulted in an error")
	assert.Contains(t, err.Error(), "unable to parse cached bundle file")
}

func TestInstallFromTagIgnoresCurrentBundle(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	err := p.Create()
	require.NoError(t, err)

	installOpts := InstallOptions{}
	installOpts.Tag = "mybun:1.0"

	err = installOpts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	assert.Empty(t, installOpts.File, "The install should ignore the bundle in the current directory because we are installing from a tag")
}

func TestBundleLifecycleOpts_ToActionArgs(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/install/base-params.json", "base-params.json")

	deps := &dependencyExecutioner{}

	t.Run("porter.yaml set", func(t *testing.T) {
		opts := BundleLifecycleOpts{}
		opts.File = "porter.yaml"
		cxt.AddTestFile("testdata/porter.yaml", "porter.yaml")
		cxt.AddTestFile("testdata/bundle.json", ".cnab/bundle.json")

		err := opts.Validate(nil, cxt.Context)
		require.NoError(t, err, "Validate failed")
		args := opts.ToActionArgs(deps)

		assert.Equal(t, ".cnab/bundle.json", args.BundlePath, "BundlePath not populated correctly")
	})

	// Just do a quick check that things are populated correctly when a bundle.json is passed
	t.Run("bundle.json set", func(t *testing.T) {
		opts := BundleLifecycleOpts{}
		opts.CNABFile = "/bundle.json"
		cxt.AddTestFile("testdata/bundle.json", "/bundle.json")

		err := opts.Validate(nil, cxt.Context)
		require.NoError(t, err, "Validate failed")
		args := opts.ToActionArgs(deps)

		assert.Equal(t, opts.CNABFile, args.BundlePath, "BundlePath was not populated correctly")
	})

	t.Run("remaining fields", func(t *testing.T) {
		opts := BundleLifecycleOpts{
			sharedOptions: sharedOptions{
				bundleFileOptions: bundleFileOptions{
					RelocationMapping: "relocation-mapping.json",
				},
				Name: "MyClaim",
				Params: []string{
					"PARAM1=VALUE1",
				},
				ParameterSets: []string{
					"HELLO_CUSTOM",
				},
				CredentialIdentifiers: []string{
					"mycreds",
				},
				Driver: "docker",
			},
			AllowAccessToDockerHost: true,
		}

		err := opts.Validate(nil, cxt.Context)
		require.NoError(t, err, "Validate failed")
		args := opts.ToActionArgs(deps)

		assert.Equal(t, opts.AllowAccessToDockerHost, args.AllowDockerHostAccess, "AllowDockerHostAccess not populated correctly")
		assert.Equal(t, opts.CredentialIdentifiers, args.CredentialIdentifiers, "CredentialIdentifiers not populated correctly")
		assert.Equal(t, opts.Driver, args.Driver, "Driver not populated correctly")
		assert.Equal(t, opts.parsedParams, args.Params, "Params not populated correctly")
		assert.Equal(t, opts.ParameterSets, args.ParameterSets, "Parameter sets not populated correctly")
		assert.Equal(t, opts.Name, args.Claim, "Claim not populated correctly")
		assert.Equal(t, opts.RelocationMapping, args.RelocationMapping, "RelocationMapping not populated correctly")
	})
}

func TestInstallFromTag_ManageFromClaim(t *testing.T) {
	p := NewTestPorter(t)

	installOpts := InstallOptions{}
	installOpts.Name = "hello"
	installOpts.Tag = "getporter/porter-hello:v0.1.0"
	err := installOpts.Validate(nil, p.Context)
	require.NoError(t, err, "InstallOptions.Validate failed")

	err = p.InstallBundle(installOpts)
	require.NoError(t, err, "InstallBundle failed")

	upgradeOpts := UpgradeOptions{}
	upgradeOpts.Name = installOpts.Name
	err = upgradeOpts.Validate(nil, p.Context)

	err = p.UpgradeBundle(upgradeOpts)
	require.NoError(t, err, "UpgradeBundle failed")

	uninstallOpts := UninstallOptions{}
	uninstallOpts.Name = installOpts.Name
	err = uninstallOpts.Validate(nil, p.Context)

	err = p.UninstallBundle(uninstallOpts)
	require.NoError(t, err, "UninstallBundle failed")
}
