package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	kahn1dot0Hash  = "887e7e65e39277f8744bd00278760b06"
	kahn1dot01     = cnab.MustParseOCIReference("deislabs/kubekahn:1.0")
	kahnlatestHash = "fd4bbe38665531d10bb653140842a370"
	kahnlatest     = cnab.MustParseOCIReference("deislabs/kubekahn:latest")
)

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
		args, err := p.BuildActionArgs(context.TODO(), claims.Installation{}, opts)
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
		args, err := p.BuildActionArgs(context.TODO(), claims.Installation{}, opts)
		require.NoError(t, err, "BuildActionArgs failed")

		assert.NotEmpty(t, args.BundleReference.Definition, "BundlePath was not populated correctly")
	})

	t.Run("remaining fields", func(t *testing.T) {
		p := NewTestPorter(t)
		p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")
		p.TestConfig.TestContext.AddTestFileFromRoot("pkg/runtime/testdata/relocation-mapping.json", "relocation-mapping.json")
		opts := InstallOptions{
			BundleActionOptions: &BundleActionOptions{
				sharedOptions: sharedOptions{
					bundleFileOptions: bundleFileOptions{
						RelocationMapping: "relocation-mapping.json",
						File:              config.Name,
					},
					Name: "MyInstallation",
					Params: []string{
						"my-first-param=1",
					},
					ParameterSets: []string{
						"porter-hello",
					},
					CredentialIdentifiers: []string{
						"mycreds",
					},
					Driver: "docker",
				},
				AllowDockerHostAccess: true,
			},
		}
		p.TestParameters.TestSecrets.AddSecret("PARAM2_SECRET", "VALUE2")
		p.TestParameters.AddTestParameters("testdata/paramset2.json")

		err := opts.Validate(nil, p.Porter)
		require.NoError(t, err, "Validate failed")
		existingInstall := claims.Installation{Name: opts.Name}
		args, err := p.BuildActionArgs(context.TODO(), existingInstall, opts)
		require.NoError(t, err, "BuildActionArgs failed")

		expectedParams := map[string]interface{}{
			"my-first-param":  1,
			"my-second-param": "VALUE2",
			"porter-debug":    true,
			"porter-state":    nil,
		}

		assert.Equal(t, opts.AllowDockerHostAccess, args.AllowDockerHostAccess, "AllowDockerHostAccess not populated correctly")
		assert.Equal(t, opts.Driver, args.Driver, "Driver not populated correctly")
		assert.EqualValues(t, expectedParams, args.Params, "Params not populated correctly")
		assert.Equal(t, existingInstall, args.Installation, "Installation not populated correctly")
		wantReloMap := relocation.ImageRelocationMap{"gabrtv/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687": "my.registry/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687"}
		assert.Equal(t, wantReloMap, args.BundleReference.RelocationMap, "RelocationMapping not populated correctly")
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

func TestBundleActionOptions_Validate(t *testing.T) {
	t.Run("allow docker host access", func(t *testing.T) {
		p := NewTestPorter(t)
		p.DataLoader = config.LoadFromEnvironment()
		p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", []byte("allow-docker-host-access: true"), 0600)
		require.NoError(t, p.Connect(context.Background()))

		opts := NewInstallOptions()
		opts.Reference = "getporter/porter-hello:v0.1.1"
		require.NoError(t, opts.Validate(nil, p.Porter))
		assert.True(t, opts.AllowDockerHostAccess)
	})

	t.Run("driver flag unset", func(t *testing.T) {
		p := NewTestPorter(t)
		p.DataLoader = config.LoadFromEnvironment()
		p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", []byte("runtime-driver: kubernetes"), 0600)
		require.NoError(t, p.Connect(context.Background()))

		opts := NewInstallOptions()
		opts.Reference = "getporter/porter-hello:v0.1.1"
		require.NoError(t, opts.Validate(nil, p.Porter))
		assert.Equal(t, "kubernetes", opts.Driver)
	})
	t.Run("driver flag set", func(t *testing.T) {
		p := NewTestPorter(t)
		p.DataLoader = config.LoadFromEnvironment()
		p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", []byte("driver: kubernetes"), 0600)
		require.NoError(t, p.Connect(context.Background()))

		opts := NewInstallOptions()
		opts.Driver = "docker"
		opts.Reference = "getporter/porter-hello:v0.1.1"
		require.NoError(t, opts.Validate(nil, p.Porter))
		assert.Equal(t, "docker", opts.Driver)
	})
}
