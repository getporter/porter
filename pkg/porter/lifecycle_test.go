package porter

import (
	"errors"
	"path"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	kahn1dot0Hash  = "887e7e65e39277f8744bd00278760b06"
	kahn1dot01     = cnab.MustParseOCIReference("deislabs/kubekahn:1.0")
	kahnlatestHash = "fd4bbe38665531d10bb653140842a370"
	kahnlatest     = cnab.MustParseOCIReference("deislabs/kubekahn:latest")
)

func TestBundlePullUpdateOpts_bundleCached(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	home, err := p.TestConfig.GetHomeDir()
	t.Logf("home dir is: %s", home)
	cacheDir, err := p.Cache.GetCacheDir()
	require.NoError(t, err, "should have had a porter cache dir")
	t.Logf("cache dir is: %s", cacheDir)
	p.TestConfig.TestContext.AddTestDirectory("testdata/cache", cacheDir)
	fullPath := filepath.Join(cacheDir, path.Join(kahn1dot0Hash, "/cnab/bundle.json"))
	fileExists, err := p.FileSystem.Exists(fullPath)
	require.True(t, fileExists, "this test requires that the file exist")
	_, ok, err := p.Cache.FindBundle(kahn1dot01)
	assert.True(t, ok, "should have found the bundle...")
	b := &BundleActionOptions{
		BundlePullOptions: BundlePullOptions{
			Reference: "deislabs/kubekahn:1.0",
		},
	}
	err = b.Validate(nil, p.Porter)
	require.NoError(t, err)
	_, err = p.prepullBundleByReference(b)
	assert.NoError(t, err, "pulling bundle should not have resulted in an error")
	assert.Equal(t, "mysql", b.Name, "name should have matched testdata bundle")
	assert.Equal(t, fullPath, b.CNABFile, "the prepare method should have set the file to the fullpath")
}

func TestBundlePullUpdateOpts_pullError(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestRegistry.MockPullBundle = func(ref cnab.OCIReference, insecureRegistry bool) (bundleReference cnab.BundleReference, err error) {
		return cnab.BundleReference{}, errors.New("unable to pull bundle deislabs/kubekahn:latest")
	}

	b := &BundleActionOptions{
		BundlePullOptions: BundlePullOptions{
			Reference: "deislabs/kubekahn:latest",
		},
	}
	err := b.Validate(nil, p.Porter)
	require.NoError(t, err)
	_, err = p.prepullBundleByReference(b)
	assert.Error(t, err, "pulling bundle should have resulted in an error")
	assert.Contains(t, err.Error(), "unable to pull bundle deislabs/kubekahn:latest")

}

func TestBundlePullUpdateOpts_cacheLies(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	// mess up the cache
	p.FileSystem.WriteFile(path.Join("/root/.porter/cache", kahn1dot0Hash, "/cnab/bundle.json"), []byte(""), 0644)

	b := &BundleActionOptions{
		BundlePullOptions: BundlePullOptions{
			Reference: "deislabs/kubekahn:1.0",
		},
	}
	err := b.Validate(nil, p.Porter)
	require.NoError(t, err)

	_, err = p.prepullBundleByReference(b)
	assert.Error(t, err, "pulling bundle should have resulted in an error")
	assert.Contains(t, err.Error(), "unable to parse cached bundle file")
}

func TestInstallFromTagIgnoresCurrentBundle(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	err := p.Create()
	require.NoError(t, err)

	installOpts := NewInstallOptions()
	installOpts.Reference = "mybun:1.0"

	err = installOpts.Validate([]string{}, p.Porter)
	require.NoError(t, err)

	assert.Empty(t, installOpts.File, "The install should ignore the bundle in the current directory because we are installing from a tag")
}

func TestPorter_BuildActionArgs(t *testing.T) {
	t.Run("no bundle set", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Teardown()
		opts := NewInstallOptions()
		opts.Name = "mybuns"

		err := opts.Validate(nil, p.Porter)
		require.Error(t, err, "Validate should fail")
		assert.Contains(t, err.Error(), "No bundle specified")
	})

	t.Run("porter.yaml set", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Teardown()

		opts := NewInstallOptions()
		opts.File = "porter.yaml"
		p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")
		p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", ".cnab/bundle.json")

		err := opts.Validate(nil, p.Porter)
		require.NoError(t, err, "Validate failed")
		args, err := p.BuildActionArgs(claims.Installation{}, opts)
		require.NoError(t, err, "BuildActionArgs failed")

		assert.NotEmpty(t, args.BundleReference.Definition)
	})

	// Just do a quick check that things are populated correctly when a bundle.json is passed
	t.Run("bundle.json set", func(t *testing.T) {
		p := NewTestPorter(t)
		opts := NewInstallOptions()
		opts.CNABFile = "/bundle.json"
		p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")

		err := opts.Validate(nil, p.Porter)
		require.NoError(t, err, "Validate failed")
		args, err := p.BuildActionArgs(claims.Installation{}, opts)
		require.NoError(t, err, "BuildActionArgs failed")

		assert.NotEmpty(t, args.BundleReference.Definition, "BundlePath was not populated correctly")
	})

	t.Run("remaining fields", func(t *testing.T) {
		p := NewTestPorter(t)
		p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")
		opts := InstallOptions{
			BundleActionOptions: &BundleActionOptions{
				sharedOptions: sharedOptions{
					bundleFileOptions: bundleFileOptions{
						RelocationMapping: "relocation-mapping.json",
						File:              config.Name,
					},
					Name: "MyInstallation",
					Params: []string{
						"PARAM1=VALUE1",
					},
					ParameterSets: []string{
						"porter-hello",
					},
					CredentialIdentifiers: []string{
						"mycreds",
					},
					Driver: "docker",
				},
				AllowAccessToDockerHost: true,
			},
		}
		p.TestParameters.TestSecrets.AddSecret("PARAM2_SECRET", "VALUE2")
		p.TestParameters.AddTestParameters("testdata/paramset2.json")

		err := opts.Validate(nil, p.Porter)
		require.NoError(t, err, "Validate failed")
		args, err := p.BuildActionArgs(claims.Installation{}, opts)
		require.NoError(t, err, "BuildActionArgs failed")

		expectedParams := map[string]string{
			"PARAM1":       "VALUE1",
			"PARAM2":       "VALUE2",
			"porter-debug": "true",
		}

		assert.Equal(t, opts.AllowAccessToDockerHost, args.AllowDockerHostAccess, "AllowDockerHostAccess not populated correctly")
		assert.Equal(t, opts.CredentialIdentifiers, args.CredentialIdentifiers, "CredentialIdentifiers not populated correctly")
		assert.Equal(t, opts.Driver, args.Driver, "Driver not populated correctly")
		assert.Equal(t, expectedParams, args.Params, "Params not populated correctly")
		assert.Equal(t, opts.Name, args.Installation, "Installation not populated correctly")
		assert.Equal(t, opts.RelocationMapping, args.BundleReference.RelocationMap, "RelocationMapping not populated correctly")
	})
}

func TestManifestIgnoredWithTag(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	t.Run("ignore manifest in cwd if tag present", func(t *testing.T) {
		opts := BundleActionOptions{}
		opts.Reference = "deislabs/kubekahn:latest"

		// `path.Join(wd...` -> makes cnab.go#defaultBundleFiles#manifestExists `true`
		// Only when `manifestExists` eq to `true`, default bundle logic will run
		p.TestConfig.TestContext.AddTestFileContents([]byte(""), config.Name)
		// When execution reach to `readFromFile`, manifest file path will be lost.
		// So, had to use root manifest file also for error simuation purpose
		p.TestConfig.TestContext.AddTestFileContents([]byte(""), config.Name)

		err := opts.Validate(nil, p.Porter)
		require.NoError(t, err, "Validate failed")
	})
}
